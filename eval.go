package boolexpr

import (
	"fmt"
)

type Symbols map[string]func() any

// Eval parses and evals the expression against a map of symbols
func Eval(s string, syms Symbols) (bool, error) {
	ast, err := Parse(s)
	if err != nil {
		return false, err
	}

	return ast.Eval(syms)
}

func (b *BoolExpr) Eval(syms Symbols) (res bool, err error) {
	res, err = b.Expr.Eval(syms)
	if err != nil {
		return
	}

	for _, e := range b.OpExprs {
		res, err = e.Eval(syms, res)
		if err != nil {
			return
		}
	}

	return
}

var ErrSymbolNotFound error

func (c Compare) Eval(syms Symbols) (res bool, err error) {
	l, err := c.Left.Eval(syms)

	return c.Right.Eval(syms, l)
}

var ErrLogicalOperationUndefinedState error

func (o *OpExpr) Eval(syms Symbols, left bool) (res bool, err error) {
	if o.Op.And {
		// short circuit if left already false
		if !left {
			return false, nil
		}

		right, err := o.Expr.Eval(syms)
		if err != nil {
			return false, err
		}

		return left && right, nil
	} else if o.Op.Or {
		// short circuit if left already true
		if left {
			return true, nil
		}

		right, err := o.Expr.Eval(syms)
		if err != nil {
			return false, err
		}

		return left || right, nil
	} else {
		return false, ErrLogicalOperationUndefinedState
	}
}

func (o *OpValue) Eval(syms Symbols, left any) (res bool, err error) {
	right, err := o.Value.Eval(syms)
	if err != nil {
		return false, err
	}

	return o.Op.Eval(left, right)
}

var ErrValueDoesntHaveAnyVal error

func (v *Value) Eval(syms Symbols) (any, error) {
	if v.Bool != nil {
		return bool(*v.Bool), nil
	} else if v.Float != nil {
		return *v.Float, nil
	} else if v.Int != nil {
		return *v.Int, nil
	} else if v.String != nil {
		return *v.String, nil
	} else if v.Ident != nil {
		sym, ok := syms[*v.Ident]
		if !ok {
			return false, ErrSymbolNotFound
		}

		return sym(), nil
	} else {
		return nil, ErrValueDoesntHaveAnyVal
	}
}

var ErrOpDoesnotHaveVal error

func (o *Op) Eval(l, r any) (res bool, err error) {
	if o.Eq {
		return o.EqEval(l, r)
	} else if o.Gt {
		return o.GtEval(l, r)
	} else if o.Gte {
		return o.GteEval(l, r)
	} else if o.Lt {
		return o.LtEval(l, r)
	} else if o.Lte {
		return o.LteEval(l, r)
	} else if o.Neq {
		return o.NeqEval(l, r)
	} else {
		return false, ErrOpDoesnotHaveVal
	}
}

var ErrorWrongDataType error

func newErrorDataTypeMismatch(op string, l, r any) error {
	return fmt.Errorf("Can't use %s on %v of type %T and %v of type %T", op, l, l, r, r)
}

func newErrorWrongDataType(op string, l any) error {
	return fmt.Errorf("Can't use %s on %v of type %T", op, l, l)
}

func (o *Op) EqEval(l, r any) (res bool, err error) {
	switch lv := l.(type) {
	case int:
		switch rv := r.(type) {
		case int:
			return lv == rv, nil
		case float64:
			return float64(lv) == rv, nil
		default:
			return false, newErrorDataTypeMismatch("=", lv, rv)
		}
	case float64:
		switch rv := r.(type) {
		case int:
			return lv == float64(rv), nil
		case float64:
			return lv == rv, nil
		default:
			return false, newErrorDataTypeMismatch("=", lv, rv)
		}
	case string:
		switch rv := r.(type) {
		case string:
			return lv == rv, nil
		default:
			return false, newErrorDataTypeMismatch("=", lv, rv)
		}
	case bool:
		switch rv := r.(type) {
		case bool:
			return lv == rv, nil
		default:
			return false, newErrorDataTypeMismatch("=", lv, rv)
		}
	default:
		return false, newErrorWrongDataType("=", lv)
	}
}

func (o *Op) GtEval(l, r any) (res bool, err error) {
	switch lv := l.(type) {
	case int:
		switch rv := r.(type) {
		case int:
			return lv > rv, nil
		case float64:
			return float64(lv) > rv, nil
		default:
			return false, newErrorDataTypeMismatch(">", lv, rv)
		}
	case float64:
		switch rv := r.(type) {
		case int:
			return lv > float64(rv), nil
		case float64:
			return lv > rv, nil
		default:
			return false, newErrorDataTypeMismatch(">", lv, rv)
		}
	case string:
		switch rv := r.(type) {
		case string:
			return lv > rv, nil
		default:
			return false, newErrorDataTypeMismatch(">", lv, rv)
		}
	default:
		return false, newErrorWrongDataType(">", lv)
	}
}
func (o *Op) GteEval(l, r any) (res bool, err error) {
	switch lv := l.(type) {
	case int:
		switch rv := r.(type) {
		case int:
			return lv >= rv, nil
		case float64:
			return float64(lv) >= rv, nil
		default:
			return false, newErrorDataTypeMismatch(">=", lv, rv)
		}
	case float64:
		switch rv := r.(type) {
		case int:
			return lv >= float64(rv), nil
		case float64:
			return lv >= rv, nil
		default:
			return false, newErrorDataTypeMismatch(">=", lv, rv)
		}
	case string:
		switch rv := r.(type) {
		case string:
			return lv >= rv, nil
		default:
			return false, newErrorDataTypeMismatch(">=", lv, rv)
		}
	default:
		return false, newErrorWrongDataType(">=", lv)
	}
}
func (o *Op) LtEval(l, r any) (res bool, err error) {
	switch lv := l.(type) {
	case int:
		switch rv := r.(type) {
		case int:
			return lv < rv, nil
		case float64:
			return float64(lv) < rv, nil
		default:
			return false, newErrorDataTypeMismatch("<", lv, rv)
		}
	case float64:
		switch rv := r.(type) {
		case int:
			return lv < float64(rv), nil
		case float64:
			return lv < rv, nil
		default:
			return false, newErrorDataTypeMismatch("<", lv, rv)
		}
	case string:
		switch rv := r.(type) {
		case string:
			return lv < rv, nil
		default:
			return false, newErrorDataTypeMismatch("<", lv, rv)
		}
	default:
		return false, newErrorWrongDataType("<", lv)
	}
}

func (o *Op) LteEval(l, r any) (res bool, err error) {
	switch lv := l.(type) {
	case int:
		switch rv := r.(type) {
		case int:
			return lv <= rv, nil
		case float64:
			return float64(lv) <= rv, nil
		default:
			return false, newErrorDataTypeMismatch("<=", lv, rv)
		}
	case float64:
		switch rv := r.(type) {
		case int:
			return lv <= float64(rv), nil
		case float64:
			return lv <= rv, nil
		default:
			return false, newErrorDataTypeMismatch("<=", lv, rv)
		}
	case string:
		switch rv := r.(type) {
		case string:
			return lv <= rv, nil
		default:
			return false, newErrorDataTypeMismatch("<=", lv, rv)
		}
	default:
		return false, newErrorWrongDataType("<=", lv)
	}
}
func (o *Op) NeqEval(l, r any) (res bool, err error) {
	switch lv := l.(type) {
	case int:
		switch rv := r.(type) {
		case int:
			return lv != rv, nil
		case float64:
			return float64(lv) != rv, nil
		default:
			return false, newErrorDataTypeMismatch("!=", lv, rv)
		}
	case float64:
		switch rv := r.(type) {
		case int:
			return lv != float64(rv), nil
		case float64:
			return lv != rv, nil
		default:
			return false, newErrorDataTypeMismatch("!=", lv, rv)
		}
	case string:
		switch rv := r.(type) {
		case string:
			return lv != rv, nil
		default:
			return false, newErrorDataTypeMismatch("!=", lv, rv)
		}
	case bool:
		switch rv := r.(type) {
		case bool:
			return lv != rv, nil
		default:
			return false, newErrorDataTypeMismatch("!=", lv, rv)
		}
	default:
		return false, newErrorWrongDataType("!=", lv)
	}
}

func (g Group) Eval(syms Symbols) (res bool, err error) {
	return g.BoolExpr.Eval(syms)
}
