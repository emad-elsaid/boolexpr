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
			stack = stack.Push(i.Expr)
			for _, e := range i.OpExprs {
				stack = stack.Push(e.Expr)
			}
		case *BoolExpr:
			stack = stack.Push(i.Expr)
			for _, e := range i.OpExprs {
				stack = stack.Push(e.Expr)
			}
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
