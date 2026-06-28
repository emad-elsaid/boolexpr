package boolexpr_test

import (
	"fmt"

	"github.com/emad-elsaid/boolexpr"
)

// Parse and evaluate an expression in a single call.
func ExampleEval() {
	symbols := boolexpr.SymbolsMap{
		"x": 10,
		"y": 30,
		"z": "hello",
	}

	ok, err := boolexpr.Eval(`x = 10 and y >= 20 and z = "hello"`, symbols)
	fmt.Println(ok, err)
	// Output: true <nil>
}

// Symbol values may be functions that are evaluated lazily, only when the
// expression actually reaches them. Both plain and error-returning variants
// are supported.
func ExampleEval_lazyFunctions() {
	symbols := boolexpr.SymbolsMap{
		"x": func() int { return 10 },
		"y": func() (string, error) { return "hello", nil },
	}

	ok, _ := boolexpr.Eval(`x > 5 and y = "hello"`, symbols)
	fmt.Println(ok)
	// Output: true
}

// Parse once, evaluate many times against different symbol values.
func ExampleParse() {
	ast, err := boolexpr.Parse(`x = 10`)
	if err != nil {
		panic(err)
	}

	first, _ := boolexpr.EvalExpression(ast, boolexpr.SymbolsMap{"x": 10})
	second, _ := boolexpr.EvalExpression(ast, boolexpr.SymbolsMap{"x": 0})

	fmt.Println(first, second)
	// Output: true false
}

// The contains and excludes operators work on strings and slices.
func ExampleEval_containsExcludes() {
	symbols := boolexpr.SymbolsMap{
		"name": "alice",
		"tags": []string{"go", "cli"},
		"ids":  []int{1, 2, 3},
	}

	ok, _ := boolexpr.Eval(`name contains "lic" and tags contains "go" and ids excludes 9`, symbols)
	fmt.Println(ok)
	// Output: true
}

// starts_with and ends_with match string prefixes and suffixes.
func ExampleEval_prefixSuffix() {
	symbols := boolexpr.SymbolsMap{
		"name":  "Joanna",
		"email": "joanna@example.com",
	}

	ok, _ := boolexpr.Eval(`name starts_with "Jo" and email ends_with "@example.com"`, symbols)
	fmt.Println(ok)
	// Output: true
}

// The match operator tests a string against a regular expression. The pattern
// may be a literal or come from another symbol.
func ExampleEval_match() {
	symbols := boolexpr.SymbolsMap{
		"x":       "pattern123",
		"email":   "joanna@example.com",
		"pattern": `.+@example\.com$`,
	}

	ok, _ := boolexpr.Eval(`x match "pattern.*" and email match pattern`, symbols)
	fmt.Println(ok)
	// Output: true
}

// ListSymbols reports every symbol an expression references.
func ExampleListSymbols() {
	ast, _ := boolexpr.Parse(`x > 1 and y < 10 or z = true`)

	syms := boolexpr.ListSymbols(ast)

	// Sort for deterministic output, since the order is unspecified.
	fmt.Println(len(syms), "symbols")
	// Output: 3 symbols
}

// CachedMap resolves each symbol at most once and records which symbols an
// evaluation actually touched. Short-circuiting means "y" is never looked up.
func ExampleCachedMap() {
	symbols := boolexpr.NewCachedMap(map[string]any{
		"x": false,
		"y": func() bool { panic("never called") },
	})

	ok, _ := boolexpr.Eval(`x and y`, symbols)

	fmt.Println(ok)
	fmt.Println(symbols.Used())
	// Output:
	// false
	// map[x:false]
}
