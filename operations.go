package boolexpr

import (
	"github.com/emad-elsaid/types"
)

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
			if i.Left.Ident != nil {
				syms = syms.Push(*i.Left.Ident)
			}
			stack = stack.Push(i.Right)
		case *OpValue:
			if i.Value.Ident != nil {
				syms = syms.Push(*i.Value.Ident)
			}
		case Group:
			stack = stack.Push(i.BoolExpr)
		case OpExpr:
			stack = stack.Push(i.Expr)
		}
	}

	return syms.Unique()
}
