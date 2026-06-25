package boolexpr

import (
	"github.com/alecthomas/participle/v2"
	"github.com/emad-elsaid/boolexpr/internal"
)

// Parse compiles the expression string s into an [Expression] tree that can be
// evaluated repeatedly with [EvalExpression]. A non-nil error is returned if s
// is not a syntactically valid expression.
func Parse(s string) (Expression, error) {
	e, error := parser.ParseString("", s)
	return Expression{e}, error
}

// parser is built once at package initialization. The grammar is static, so a
// build failure is a programming error; MustBuild panics immediately with a
// clear message rather than leaving a nil parser to nil-deref on first Parse.
var parser = participle.MustBuild[internal.BoolExpr](
	participle.Unquote("String"),
	participle.Union[internal.Expr](internal.Compare{}, internal.SubExpr{}, internal.BoolValue{}),
)

// Expression is a parsed boolean expression tree produced by [Parse]. It holds
// no symbol values and can be evaluated repeatedly, and concurrently, against
// different [Symbols] using [EvalExpression]. The zero value is not usable;
// always obtain an Expression from [Parse].
type Expression struct {
	e *internal.BoolExpr
}
