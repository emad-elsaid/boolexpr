// Package boolexpr parses and evaluates boolean expressions against a set of
// named variables (called "symbols").
//
// An expression is a human-readable string such as:
//
//	x == 0 or (y > 20 and z = "hello")
//
// It is parsed once into an [Expression] tree and can then be evaluated many
// times against different symbol values, returning a single bool.
//
// # Quick start
//
// Parse and evaluate in a single call with [Eval]:
//
//	symbols := boolexpr.SymbolsMap{
//		"x": 10,
//		"y": func() int { return 30 },
//		"z": func() (string, error) { return "hello", nil },
//	}
//	ok, err := boolexpr.Eval(`x = 10 and y >= 20 and z = "hello"`, symbols)
//	// ok == true, err == nil
//
// To evaluate the same expression repeatedly, parse it once with [Parse] and
// reuse the result with [EvalExpression]:
//
//	ast, err := boolexpr.Parse(`x = 10`)
//	ok1, _ := boolexpr.EvalExpression(ast, boolexpr.SymbolsMap{"x": 10})
//	ok2, _ := boolexpr.EvalExpression(ast, boolexpr.SymbolsMap{"x": 0})
//	// ok1 == true, ok2 == false
//
// # Syntax
//
// A comparison always takes the form "value operator value", where each value
// is either a literal or a symbol name:
//
//	x > 1
//	name = "alice"
//	count <= other_count
//
// Supported comparison operators:
//
//	=  ==  !=  >  <  >=  <=  contains  excludes  starts_with  ends_with  match
//
// Comparisons are joined with the logical operators "and" (or "&&") and "or"
// (or "||"), and may be grouped with parentheses:
//
//	x > 10 and y < 20 or (z = true and active)
//
// A bare boolean symbol or literal may be used without a comparison operator,
// e.g. "active" or "true".
//
// Literal value types are int, float, string and bool. Strings are written
// with double quotes.
//
// # Symbols
//
// A [Symbols] provides the value for each symbol name during evaluation.
// [SymbolsMap] is the simplest implementation, wrapping a map[string]any.
//
// A value may be a literal (string, int, float64, bool), a slice
// ([]string, []int, []float64, []bool) for use with contains/excludes, or a
// function that is called lazily during evaluation. Both plain
// (func() int) and error-returning (func() (int, error)) function variants are
// supported for every value type; an error returned by a function aborts
// evaluation. See [resolveSymbol] for the full list of accepted function
// signatures.
//
// Functions are only called when evaluation actually reaches the symbol, so
// expensive lookups can be deferred and skipped via short-circuiting.
//
// # contains and excludes
//
// "contains" tests whether the left operand contains the right operand;
// "excludes" is its negation.
//
//	"hello" contains "ell"   // substring match on strings
//	tags contains "go"       // element match when tags is a []string
//	ids excludes 0           // element match when ids is a []int / []float64
//
// For numeric slices int and float64 are interchangeable, matching the
// behaviour of "=".
//
// # starts_with and ends_with
//
// Both operands must be strings:
//
//	name starts_with "Jo"
//	email ends_with "@example.com"
//
// # match
//
// "match" tests the left string against a regular expression (Go's RE2 syntax)
// given as the right operand. The pattern may be a literal or another symbol:
//
//	x match "pattern.*"
//	email match valid_email_regex
//
// Both operands must be strings and the match is unanchored. Compiled patterns
// are cached, so each distinct pattern is compiled only once across all
// evaluations.
//
// # Short-circuit evaluation
//
// Logical operators short-circuit: with "and" a false left operand skips the
// right operand (and any symbol functions it would call); with "or" a true
// left operand does the same. Combined with [SymbolsCached], this lets you
// inspect exactly which symbols an evaluation actually touched.
package boolexpr
