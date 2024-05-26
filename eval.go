package boolexpr

import (
	"fmt"
)

type Symbols map[string]func() any

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
	sym, ok := syms[c.Left]
	if !ok {
		return false, ErrSymbolNotFound
	}

	val := sym()
	return c.Right.Eval(syms, val)
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
	right, err := o.Value.Eval()
	if err != nil {
		return false, err
	}

	return o.Op.Eval(left, right)
}

var ErrValueDoesntHaveAnyVal error

func (v *Value) Eval() (any, error) {
	if v.Bool != nil {
		return *v.Bool, nil
	} else if v.Float != nil {
		return *v.Float, nil
	} else if v.Int != nil {
		return *v.Int, nil
	} else if v.String != nil {
		return *v.String, nil
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

func (o *Op) EqEval(l, r any) (res bool, err error) {
	switch lv := l.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64:
		switch rv := r.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64:
			return lv == rv, nil
		default:
			return false, fmt.Errorf("Can't use = on %#v, %#v", lv, rv)
		}
	case string:
		switch rv := r.(type) {
		case string:
			return lv == rv, nil
		default:
			return false, fmt.Errorf("Can't use = on %#v, %#v", lv, rv)
		}
	case bool:
		switch rv := r.(type) {
		case bool:
			return lv == rv, nil
		default:
			return false, fmt.Errorf("Can't use = on %#v, %#v", lv, rv)
		}
	default:
		return false, fmt.Errorf("Can't use = on type %#v", lv)
	}
}

func (o *Op) GtEval(l, r any) (res bool, err error) {
	switch lv := l.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64:
		switch rv := r.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64:
			return lv.(float64) > rv.(float64), nil
		default:
			return false, fmt.Errorf("Can't use > on %#v, %#v", lv, rv)
		}
	case string:
		switch rv := r.(type) {
		case string:
			return lv > rv, nil
		default:
			return false, fmt.Errorf("Can't use > on %#v, %#v", lv, rv)
		}
	default:
		return false, fmt.Errorf("Can't use > on type %#v", lv)
	}
}
func (o *Op) GteEval(l, r any) (res bool, err error) {
	switch lv := l.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64:
		switch rv := r.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64:
			return lv.(float64) >= rv.(float64), nil
		default:
			return false, fmt.Errorf("Can't use >= on %#v, %#v", lv, rv)
		}
	case string:
		switch rv := r.(type) {
		case string:
			return lv >= rv, nil
		default:
			return false, fmt.Errorf("Can't use >= on %#v, %#v", lv, rv)
		}
	default:
		return false, fmt.Errorf("Can't use >= on type %#v", lv)
	}
}
func (o *Op) LtEval(l, r any) (res bool, err error) {
	switch lv := l.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64:
		switch rv := r.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64:
			return lv.(float64) < rv.(float64), nil
		default:
			return false, fmt.Errorf("Can't use < on %#v, %#v", lv, rv)
		}
	case string:
		switch rv := r.(type) {
		case string:
			return lv < rv, nil
		default:
			return false, fmt.Errorf("Can't use < on %#v, %#v", lv, rv)
		}
	default:
		return false, fmt.Errorf("Can't use < on type %#v", lv)
	}
}
func (o *Op) LteEval(l, r any) (res bool, err error) {
	switch lv := l.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64:
		switch rv := r.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64:
			return lv.(float64) <= rv.(float64), nil
		default:
			return false, fmt.Errorf("Can't use <= on %#v, %#v", lv, rv)
		}
	case string:
		switch rv := r.(type) {
		case string:
			return lv <= rv, nil
		default:
			return false, fmt.Errorf("Can't use <= on %#v, %#v", lv, rv)
		}
	default:
		return false, fmt.Errorf("Can't use <= on type %#v", lv)
	}
}
func (o *Op) NeqEval(l, r any) (res bool, err error) {
	switch lv := l.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64:
		switch rv := r.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64:
			return lv != rv, nil
		default:
			return false, fmt.Errorf("Can't use != on %#v, %#v", lv, rv)
		}
	case string:
		switch rv := r.(type) {
		case string:
			return lv != rv, nil
		default:
			return false, fmt.Errorf("Can't use != on %#v, %#v", lv, rv)
		}
	case bool:
		switch rv := r.(type) {
		case bool:
			return lv != rv, nil
		default:
			return false, fmt.Errorf("Can't use != on %#v, %#v", lv, rv)
		}
	default:
		return false, fmt.Errorf("Can't use != on type %#v", lv)
	}
}

func (g Group) Eval(syms Symbols) (res bool, err error) {
	return g.BoolExpr.Eval(syms)
}
