# BoolExpr

[![Go Reference](https://pkg.go.dev/badge/github.com/emad-elsaid/boolexpr.svg)](https://pkg.go.dev/github.com/emad-elsaid/boolexpr)
[![Go Report Card](https://goreportcard.com/badge/github.com/emad-elsaid/boolexpr)](https://goreportcard.com/report/github.com/emad-elsaid/boolexpr)
[![codecov](https://codecov.io/gh/emad-elsaid/boolexpr/graph/badge.svg?token=QBXTR1XRD6)](https://codecov.io/gh/emad-elsaid/boolexpr)

A Go package to evaluate boolean expressions. against a map of variables. variables can be lazy computed using a function that return the value.

BoolExpr allows your program user to write a bool expression in the form: `x = 10 and y > 20 and z = "hello"` and then run it many times against functions: `x,y,z` which returns `int, int, string` values. the evaluation returns a simple `bool`

# Usage

You can parse and evaluate the expression in one call
```go
exp := `x = 10 and y >= 20 and z = "hello"`
symbols := map[string]func() any{
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

symbols := map[string]func() any{
    "x": func() any { return 10 },
    "y": func() any { return 30 },
    "z": func() any { return "hello" },
}
output, err := ast.Eval(symbols) // Output: true, nil

symbols = map[string]func() any{
    "x": func() any { return 0 },
    "y": func() any { return 30 },
    "z": func() any { return "hello" },
}
output, err = ast.Eval(symbols) // Output: false, nil
```

# Syntax

The syntax supports:

* The following comparisons: =, !=, >, <, >=, <=
* And the logical operators: and, or
* And the values types: int, float, string, bool
* logical expressions can be grouped with `(...)`
* The comparison must always be in the form `value operator value`
  * value can be a symbol or a literal e.g `x`, `1`, `true`, `"hello"`
  * operator is one of the comparison operators
Symbols map is a map from `string` (the variable name) to `any` value:
* If the value is a literal (string, int, float, bool) it'll be used
* If it's a `func() string/int/float/bool` it'll be evaluated and the return value will be used
* If it's a `func() any` it'll be also evaluated and the return value used.
* If it's a `func() (string/int/float/bool, error)` the value returned will be used if no error. If an error is returned the evaluation is terminated and the error is returned.

# Expressions examples:

* `x > 1`
* `x > y and y > z`
* `x = 10 and y != 20`
* `x > 10 and y < 20 or z = true`
* `x != 20 or y = 30 or z = "helloworld" or (a = false and b = true)`

# Evaluation

BoolExpr will short circuit in two situations:

* If `and` is used and the left operand is `false`, the right operand will not be executed and it'll return `false`
* If `or` is used and the left operand is `true`, the right opreand will not be executed and it'll return `true`

# Examples

* A basic example in Go playground: https://go.dev/play/p/gONexex0d7Q
