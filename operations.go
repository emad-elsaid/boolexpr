package boolexpr

import (
	. "github.com/emad-elsaid/boolexpr/internal"
	"github.com/emad-elsaid/types"
)

// ListSymbols returns the unique symbol names referenced anywhere in the parsed
// expression, in no particular order. It is useful for validating that every
// required symbol is available before evaluation, or for building the Symbols
// set on demand.
func ListSymbols(exp Expression) []string {
	var stack types.Slice[any]
	stack = stack.Push(exp.e)
	syms := types.Slice[string]{}

	for len(stack) > 0 {
		var item any
		stack, item = stack.Pop()

		switch i := item.(type) {
		case BoolExpr:
			stack = pushBoolExpr(stack, i)
		case *BoolExpr:
			stack = pushBoolExpr(stack, *i)
		case Compare:
			stack = stack.Push(i.Left)
			stack = stack.Push(i.Right)
		case BoolValue:
			stack = stack.Push(i.Value)
		case Value:
			if i.Symbol != nil {
				syms = syms.Push(*i.Symbol)
			}
		case SubExpr:
			stack = stack.Push(i.BoolExpr)
		}
	}

	return syms.Unique()
}

// pushBoolExpr pushes every primary expression of a BoolExpr — the leading
// AND-expression and each OR-ed AND-expression, including all of their AND-ed
// operands — onto the walk stack.
func pushBoolExpr(stack types.Slice[any], b BoolExpr) types.Slice[any] {
	stack = pushAndExpr(stack, b.And)
	for _, o := range b.OrOps {
		stack = pushAndExpr(stack, o.And)
	}

	return stack
}

func pushAndExpr(stack types.Slice[any], a AndExpr) types.Slice[any] {
	stack = stack.Push(a.Expr)
	for _, op := range a.AndOps {
		stack = stack.Push(op.Expr)
	}

	return stack
}
