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

# Sytax

The syntax supports:

* The following comparisons: =, !=, >, <, >=, <=
* And the logical operators: and, or
* And the values types: int, float, string, bool
* The comparison must always be in the form `Symbol operator literal`
  * Where `Symbol` is a variable name in the symbols map
  * operator is one of the comparison operators
  * literal is one of the values types e.g: 1, 3.14, "Hello", true/false
