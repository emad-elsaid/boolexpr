package boolexpr

import (
	"github.com/alecthomas/participle/v2"
	"github.com/emad-elsaid/boolexpr/internal"
)

// Parse will convert string s to BoolExpr tree that can be evaluated multiple times
func Parse(s string) (Expression, error) {
	e, error := parser.ParseString("", s)
	return Expression{e}, error
}

var parser, parserErr = participle.Build[internal.BoolExpr](
	participle.Unquote("String"),
	participle.Union[internal.Expr](internal.Compare{}, internal.SubExpr{}),
)

type Expression struct {
	e *internal.BoolExpr
}
