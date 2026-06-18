# BoolExpr

[![Go Reference](https://pkg.go.dev/badge/github.com/emad-elsaid/boolexpr.svg)](https://pkg.go.dev/github.com/emad-elsaid/boolexpr)
[![Go Report Card](https://goreportcard.com/badge/github.com/emad-elsaid/boolexpr)](https://goreportcard.com/report/github.com/emad-elsaid/boolexpr)
[![codecov](https://codecov.io/gh/emad-elsaid/boolexpr/graph/badge.svg?token=QBXTR1XRD6)](https://codecov.io/gh/emad-elsaid/boolexpr)

A Go package to evaluate boolean expressions. against a map of variables. variables can be lazy computed using a function that return the value.

BoolExpr allows your program user to write a bool expression in the form: `w == 0 or x = 10 and y > 20 and z = "hello"` and then run it many times against functions: `x,y,z` which returns `int, int, string` values. the evaluation returns a simple `bool`

# Usage

You can parse and evaluate the expression in one call
```go
exp := `x = 10 and y >= 20 and z = "hello"`
var symbols SymbolsMap = map[string]any{
    "x": func() any { return 10 },
    "y": 30,
    "z": func() string { return "hello" },
}
output, err := Eval(exp, symbols) // Output: true, nil
```

or you can parse and evaulate multiple times

```go
exp := `x = 10 and y >= 20 and z = "hello"`
ast, err := Parse(exp)

var symbols SymbolsMap = map[string]any{
    "x": func() any { return 10 },
    "y": func() any { return 30 },
    "z": func() any { return "hello" },
}
output, err := EvalBoolExpr(ast, symbols) // Output: true, nil

symbols = map[string]func() any{
    "x": func() any { return 0 },
    "y": func() any { return 30 },
    "z": func() any { return "hello" },
}
output, err = EvalBoolExpr(ast, symbols) // Output: false, nil
```

# Syntax

The syntax supports:

* The following comparisons: `=`, `==`, `!=`, `>`, `<`, `>=`, `<=`, `contains`, `excludes`, `starts_with`, `ends_with`, `match`
* And the logical operators: `and` (or `&&`), `or` (or `||`)
* And the values types: int, float, string, bool
* logical expressions can be grouped with `(...)`
* The comparison must always be in the form `value operator value`
  * value can be a symbol or a literal e.g `x`, `1`, `true`, `"hello"`
  * operator is one of the comparison operators
* A bare bool symbol or literal can be used without a comparison operator e.g. `active`, `true`

Symbols map is a map from `string` (the variable name) to `any` value:
* If the value is a literal (string, int, float, bool) it'll be used
* If it's a `func() string/int/float/bool` it'll be evaluated and the return value will be used
* If it's a `func() any` it'll be also evaluated and the return value used.
* If it's a `func() (string/int/float/bool, error)` the value returned will be used if no error. If an error is returned the evaluation is terminated and the error is returned.
* If it's a `[]string`, `[]int`, `[]float64`, or `[]bool` it can be used with the `contains`/`excludes` operators.
* The func variants `func() []T` and `func() ([]T, error)` are also supported for each slice type.

### The `contains` and `excludes` operators

`contains` tests whether the left operand contains the right operand. `excludes` is its negation.

| Left type   | Right type         | Behaviour                                               |
|-------------|--------------------|---------------------------------------------------------|
| `string`    | `string`           | substring match (`strings.Contains`)                    |
| `[]string`  | `string`           | element equality                                        |
| `[]int`     | `int` or `float64` | element equality (int↔float64 compatible, same as `=`)  |
| `[]float64` | `float64` or `int` | element equality (int↔float64 compatible)               |
| `[]bool`    | `bool`             | element equality                                        |

Type mismatches (e.g. `[]string contains 1`) return an error.

### The `starts_with` and `ends_with` operators

Both operands must be `string`. Returns an error for any other type.

| Expression | Behaviour |
|---|---|
| `"hello" starts_with "he"` | `strings.HasPrefix` |
| `"hello" ends_with "lo"` | `strings.HasSuffix` |

### The `match` operator

`match` tests whether the left string matches the regular expression given as
the right operand. Both operands must be `string`; any other type returns an
error. The pattern may be a literal or another symbol that resolves to a string.

The pattern uses Go's [regexp](https://pkg.go.dev/regexp) (RE2) syntax and the
match is unanchored (use `^`/`$` to anchor). Compiled patterns are cached, so
each distinct pattern is compiled only once across all evaluations.

| Expression | Behaviour |
|---|---|
| `x match "pattern.*"` | matches `x` against the literal pattern |
| `email match pattern` | matches `email` against the regex held in symbol `pattern` |

# Expressions examples:

* `x > 1`
* `x > y and y > z`
* `x = 10 and y != 20`
* `x > 10 and y < 20 or z = true`
* `x != 20 or y = 30 or z = "helloworld" or (a = false and b = true)`
* `name contains "alice"`
* `tags contains "go"`
* `roles contains "admin" and active = true`
* `tags excludes "deprecated"`
* `ids excludes 0 and active = true`
* `name starts_with "Jo"`
* `email ends_with "@example.com"`
* `name match "^[A-Z][a-z]+$"`
* `email match valid_email_regex`

# Evaluation

BoolExpr will short circuit in two situations:

* If `and` is used and the left operand is `false`, the right operand will not be executed and it'll return `false`
* If `or` is used and the left operand is `true`, the right opreand will not be executed and it'll return `true`

# Examples

* A basic example in Go playground: https://go.dev/play/p/4mr_z20q3C2
