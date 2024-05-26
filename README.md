# BoolExpr

A Go package to evaluate boolean expressions. against a map of variables. variables are lazy computed using a function that return the value.

BoolExpr allows your program user to write a bool expression in the form: `x = 10 and y > 20 and z = "hello"` and then run it many times agains function `x,y,z` which returns `int, int, string` values. the evaluation returns a simple `bool`

# Usage

You can parse and evaluate the expression in one call
```go
exp := `x = 10 and y >= 20 and z = "hello"`
symbols := map[string]func() any{
    "x": func() any { return 10 },
    "y": func() any { return 30 },
    "z": func() any { return "hello" },
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
* The comparison must always be in the form `Symbol operator literal`
  * Where `Symbol` is a variable name in the symbols map
  * operator is one of the comparison operators
  * literal is one of the values types e.g: 1, 3.14, "Hello", true/false

# Expressions examples:

* `x > 1`
* `x = 10 and y != 20`
* `x > 10 and y < 20 or z = true`
* `x != 20 or y = 30 or z = "helloworld" or (a = false and b = true)`

# Evaluation

BoolExpr will short circuit in two situations:

* If `and` is used and the left operand is `false`, the right operand will not be executed and it'll return `false`
* If `or` is used and the left operand is `true`, the right opreand will not be executed and it'll return `true`

# Examples

* A basic example in Go playground: https://go.dev/play/p/gONexex0d7Q
