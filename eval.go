package boolexpr

import (
	"errors"
	"fmt"
)

var (
	ErrSymbolNotFound                 = errors.New("Symbol not found")
	ErrLogicalOperationUndefinedState = errors.New("Logical operation not specified")
	ErrValueDoesntHaveAnyVal          = errors.New("Value is not specified")
	ErrSymbolTypeUnknown              = errors.New("Symbol type not supported")
	ErrOpDoesnotHaveVal               = errors.New("Operation not specified")
	ErrorWrongDataType                = errors.New("Wrong data type")
)

func newErrorDataTypeMismatch(op string, l, r any) error {
	return fmt.Errorf("%w, Can't use %s on %v of type %T and %v of type %T",
		ErrorWrongDataType,
		op, l, l, r, r)
}

func newErrorWrongDataType(op string, l any) error {
	return fmt.Errorf("%w, Can't use %s on %v of type %T",
		ErrorWrongDataType,
		op, l, l)
}

type Symbols map[string]any

// Eval parses and evals the expression against a map of symbols
func Eval(s string, syms Symbols) (bool, error) {
	ast, err := Parse(s)
	if err != nil {
		return false, err
	}

	return EvalBoolExpr(ast, syms)
}

// EvalBoolExpr evaluate a parsed expression against a map of symbols
func EvalBoolExpr(b *BoolExpr, syms Symbols) (res bool, err error) {
	if res, err = evalExpr(b.Expr, syms); err != nil {
		return
	}

	for _, e := range b.OpExprs {
		res, err = evalOpExpr(e, syms, res)
		if err != nil {
			return
		}
	}

	return
}

func evalExpr(b Expr, syms Symbols) (res bool, err error) {
	switch e := b.(type) {
	case Compare:
		l, err := evalValue(e.Left, syms)
		if err != nil {
			return false, err
		}

		r, err := evalValue(e.Right, syms)
		if err != nil {
			return false, err
		}

		return evalComparisonOp(e.Op, l, r)
	case SubExpr:
		if res, err = EvalBoolExpr(&e.BoolExpr, syms); err != nil {
			return
		}
	default:
		return false, fmt.Errorf("Expr type is unhandled %T", b)
	}

	return
}

func evalOpExpr(o OpExpr, syms Symbols, left bool) (res bool, err error) {
	if o.Op.And {
		// short circuit if left already false
		if !left {
			return false, nil
		}

		right, err := evalExpr(o.Expr, syms)
		if err != nil {
			return false, err
		}

		return left && right, nil
	} else if o.Op.Or {
		// short circuit if left already true
		if left {
			return true, nil
		}

		right, err := evalExpr(o.Expr, syms)
		if err != nil {
			return false, err
		}

		return left || right, nil
	} else {
		return false, ErrLogicalOperationUndefinedState
	}
}

func evalValue(v Value, syms Symbols) (any, error) {
	if v.Bool != nil {
		return bool(*v.Bool), nil
	} else if v.Float != nil {
		return *v.Float, nil
	} else if v.Int != nil {
		return *v.Int, nil
	} else if v.String != nil {
		return *v.String, nil
	} else if v.Symbol != nil {
		sym, ok := syms[*v.Symbol]
		if !ok {
			return false, fmt.Errorf("%w, Symbol: %s", ErrSymbolNotFound, *v.Symbol)
		}

		switch i := sym.(type) {
		case bool:
			return i, nil
		case func() bool:
			return i(), nil
		case func() (bool, error):
			return i()
		case int:
			return i, nil
		case func() int:
			return i(), nil
		case func() (int, error):
			return i()
		case string:
			return i, nil
		case func() string:
			return i(), nil
		case func() (string, error):
			return i()
		case float64:
			return i, nil
		case func() float64:
			return i(), nil
		case func() (float64, error):
			return i()
		case func() any:
			return i(), nil
		case func() (any, error):
			return i()
		default:
			return false, fmt.Errorf("%w, Symbol: %s of type %T", ErrSymbolTypeUnknown, *v.Symbol, i)
		}
	} else {
		return nil, ErrValueDoesntHaveAnyVal
	}
}

func evalComparisonOp(o ComparisonOp, l, r any) (res bool, err error) {
	if o.Eq {
		return eqEval(l, r)
	} else if o.Gt {
		return gtEval(l, r)
	} else if o.Gte {
		return gteEval(l, r)
	} else if o.Lt {
		return ltEval(l, r)
	} else if o.Lte {
		return lteEval(l, r)
	} else if o.Neq {
		return neqEval(l, r)
	} else {
		return false, ErrOpDoesnotHaveVal
	}
}

func eqEval(l, r any) (res bool, err error) {
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

func gtEval(l, r any) (res bool, err error) {
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

func gteEval(l, r any) (res bool, err error) {
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

func ltEval(l, r any) (res bool, err error) {
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

func lteEval(l, r any) (res bool, err error) {
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
func neqEval(l, r any) (res bool, err error) {
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
