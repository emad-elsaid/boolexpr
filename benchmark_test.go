package boolexpr

import (
	"fmt"
	"testing"
)

// The benchmarks below profile the three phases independently — parsing,
// evaluation, and symbol resolution — plus a few representative end-to-end
// workloads. Run with:
//
//	go test -run=^$ -bench=. -benchmem
//
// To capture profiles for later optimization:
//
//	go test -run=^$ -bench=BenchmarkEval -benchmem \
//		-cpuprofile=cpu.out -memprofile=mem.out
//	go tool pprof cpu.out

// benchSink prevents the compiler from optimizing away benchmarked results.
var (
	benchBool bool
	benchErr  error
	benchExpr Expression
	benchSyms []string
)

// ---------------------------------------------------------------------------
// Parsing
// ---------------------------------------------------------------------------

// parseCases covers expressions of growing structural complexity so we can see
// how parse cost scales with operator count, grouping, and value variety.
var parseCases = []struct {
	name string
	expr string
}{
	{"Simple", `x = 1`},
	{"Comparison", `x >= 100`},
	{"StringEq", `name = "hello world"`},
	{"And", `x = 1 and y = 2`},
	{"Or", `x = 1 or y = 2`},
	{"Mixed", `x = 10 and y >= 20 and z = "hello"`},
	{"Grouped", `x != 20 or y = 30 or (a = false and b = true)`},
	{"Deep", `(((a = 1 and b = 2) or c = 3) and (d = 4 or e = 5)) and f = 6`},
	{"Operators", `name contains "go" and tag excludes "x" and s starts_with "a" and e ends_with "z"`},
	{"Match", `email match ".+@example\\.com$"`},
}

func BenchmarkParse(b *testing.B) {
	for _, tc := range parseCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			var (
				e   Expression
				err error
			)
			for i := 0; i < b.N; i++ {
				e, err = Parse(tc.expr)
			}
			benchExpr, benchErr = e, err
		})
	}
}

// ---------------------------------------------------------------------------
// Evaluation of a pre-parsed expression (the hot path for reused expressions)
// ---------------------------------------------------------------------------

// evalCases pair an expression with matching symbols, isolating evaluation cost
// from parsing. Each focuses on one value type or operator family.
var evalCases = []struct {
	name string
	expr string
	syms SymbolsMap
}{
	{"IntLiteral", `x = 1`, SymbolsMap{"x": 1}},
	{"IntCompare", `x >= 100`, SymbolsMap{"x": 150}},
	{"FloatCompare", `x <= 3.14`, SymbolsMap{"x": 2.71}},
	{"StringEq", `s = "hello"`, SymbolsMap{"s": "hello"}},
	{"BoolBare", `active`, SymbolsMap{"active": true}},
	{"And", `x = 1 and y = 2`, SymbolsMap{"x": 1, "y": 2}},
	{"Or", `x = 1 or y = 2`, SymbolsMap{"x": 9, "y": 2}},
	{"Mixed", `x = 10 and y >= 20 and z = "hello"`, SymbolsMap{"x": 10, "y": 30, "z": "hello"}},
	{"Grouped", `x != 20 or (a = false and b = true)`, SymbolsMap{"x": 5, "a": false, "b": true}},
	{"ContainsString", `s contains "ell"`, SymbolsMap{"s": "hello"}},
	{"ContainsSlice", `tags contains "go"`, SymbolsMap{"tags": []string{"c", "go", "rust"}}},
	{"StartsWith", `s starts_with "he"`, SymbolsMap{"s": "hello"}},
	{"EndsWith", `s ends_with "lo"`, SymbolsMap{"s": "hello"}},
	{"Match", `s match "h.*o"`, SymbolsMap{"s": "hello"}},
	{"FuncSymbol", `x = 1`, SymbolsMap{"x": func() int { return 1 }}},
	{"FuncSymbolErr", `x = 1`, SymbolsMap{"x": func() (int, error) { return 1, nil }}},
}

func BenchmarkEval(b *testing.B) {
	for _, tc := range evalCases {
		ast, err := Parse(tc.expr)
		if err != nil {
			b.Fatalf("parse %q: %v", tc.expr, err)
		}

		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			var (
				res bool
				e   error
			)
			for i := 0; i < b.N; i++ {
				res, e = EvalExpression(ast, tc.syms)
			}
			benchBool, benchErr = res, e
		})
	}
}

// ---------------------------------------------------------------------------
// End-to-end: parse + evaluate together (the Eval convenience path)
// ---------------------------------------------------------------------------

func BenchmarkEvalParseAndRun(b *testing.B) {
	expr := `x = 10 and y >= 20 and z = "hello"`
	syms := SymbolsMap{"x": 10, "y": 30, "z": "hello"}

	b.ReportAllocs()
	var (
		res bool
		err error
	)
	for i := 0; i < b.N; i++ {
		res, err = Eval(expr, syms)
	}
	benchBool, benchErr = res, err
}

// ---------------------------------------------------------------------------
// Symbols implementations: SymbolsMap vs SymbolsCached
// ---------------------------------------------------------------------------

// BenchmarkSymbols compares the two Symbols implementations under repeated
// evaluation. SymbolsCached resolves each function once; SymbolsMap re-resolves
// on every lookup, so this highlights the trade-off for expensive symbols.
func BenchmarkSymbols(b *testing.B) {
	expr := `a = 1 and b = 2 and c = 3 and d = 4`
	ast, err := Parse(expr)
	if err != nil {
		b.Fatal(err)
	}

	raw := map[string]any{
		"a": func() int { return 1 },
		"b": func() int { return 2 },
		"c": func() int { return 3 },
		"d": func() int { return 4 },
	}

	b.Run("SymbolsMap", func(b *testing.B) {
		syms := SymbolsMap(raw)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchBool, benchErr = EvalExpression(ast, syms)
		}
	})

	b.Run("SymbolsCached", func(b *testing.B) {
		syms := NewSymbolsCached(raw)
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchBool, benchErr = EvalExpression(ast, syms)
		}
	})
}

// BenchmarkSymbolsConcurrent exercises SymbolsCached under concurrent access,
// the scenario it is designed for.
func BenchmarkSymbolsConcurrent(b *testing.B) {
	ast, err := Parse(`a = 1 and b = 2 and c = 3`)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		syms := NewSymbolsCached(map[string]any{
			"a": func() int { return 1 },
			"b": func() int { return 2 },
			"c": func() int { return 3 },
		})
		for pb.Next() {
			benchBool, benchErr = EvalExpression(ast, syms)
		}
	})
}

// ---------------------------------------------------------------------------
// match operator: cache hit vs the work avoided by caching
// ---------------------------------------------------------------------------

// BenchmarkMatch measures the steady-state cost of `match` once the pattern is
// cached (the common case for a reused expression).
func BenchmarkMatch(b *testing.B) {
	ast, err := Parse(`s match "^[a-z]+@example\\.com$"`)
	if err != nil {
		b.Fatal(err)
	}
	syms := SymbolsMap{"s": "joanna@example.com"}

	// Warm the pattern cache so we measure the hit path.
	if _, err := EvalExpression(ast, syms); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		benchBool, benchErr = EvalExpression(ast, syms)
	}
}

// ---------------------------------------------------------------------------
// ListSymbols: tree walk cost
// ---------------------------------------------------------------------------

func BenchmarkListSymbols(b *testing.B) {
	ast, err := Parse(`a > 1 and b < 2 or (c = 3 and d != 4) or e contains "x"`)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	var syms []string
	for i := 0; i < b.N; i++ {
		syms = ListSymbols(ast)
	}
	benchSyms = syms
}

// ---------------------------------------------------------------------------
// Scaling: evaluation cost as the number of AND-ed clauses grows
// ---------------------------------------------------------------------------

func BenchmarkEvalScaling(b *testing.B) {
	for _, n := range []int{1, 4, 16, 64} {
		expr, syms := buildConjunction(n)
		ast, err := Parse(expr)
		if err != nil {
			b.Fatalf("parse n=%d: %v", n, err)
		}

		b.Run(fmt.Sprintf("Clauses=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				benchBool, benchErr = EvalExpression(ast, syms)
			}
		})
	}
}

// buildConjunction creates an expression of n "vK = K" clauses joined by "and",
// all true, so evaluation must visit every clause (no short-circuit).
func buildConjunction(n int) (string, SymbolsMap) {
	syms := make(SymbolsMap, n)
	expr := ""
	for i := 0; i < n; i++ {
		if i > 0 {
			expr += " and "
		}
		key := fmt.Sprintf("v%d", i)
		expr += fmt.Sprintf("%s = %d", key, i)
		syms[key] = i
	}
	return expr, syms
}
