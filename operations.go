package boolexpr

import (
	"github.com/emad-elsaid/types"
)

// ListSymbols returns a list of symbols used in the expression
func ListSymbols(exp *BoolExpr) []string {
	var stack types.Slice[any]
	stack = stack.Push(exp)
	syms := types.Slice[string]{}

	for len(stack) > 0 {
		var item any
		stack, item = stack.Pop()

		switch i := item.(type) {
		case *BoolExpr:
			stack = stack.Push(i.Expr)
			for _, e := range i.OpExprs {
				stack = stack.Push(e.Expr)
			}
		case Compare:
			stack = stack.Push(i.Left)
			stack = stack.Push(i.Right)
		case *OpValue:
			stack = stack.Push(i.Value)
		case *Value:
			if i.Ident != nil {
				syms = syms.Push(*i.Ident)
			}
		case Group:
			stack = stack.Push(i.BoolExpr)
		}
	}

	return syms.Unique()
}
