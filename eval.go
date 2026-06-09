package boolexpr

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	. "github.com/emad-elsaid/boolexpr/internal"
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

// Eval parses and evals the expression against a map of symbols
func Eval(s string, syms Symbols) (bool, error) {
	ast, err := Parse(s)
	if err != nil {
		return false, err
	}

	return EvalExpression(ast, syms)
}

// EvalExpression evaluate a parsed expression against a map of symbols
func EvalExpression(e Expression, syms Symbols) (res bool, err error) {
	if res, err = evalExpr(e.e.Expr, syms); err != nil {
		return
	}

	for _, e := range e.e.OpExprs {
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

		return evalComparisonOpVal(e.Op, l, r)
	case BoolValue:
		v, err := evalValue(e.Value, syms)
		if err != nil {
			return false, err
		}

		bv, ok := v.toBool()
		if !ok {
			return false, fmt.Errorf("%w, bare value must be bool, got %T", ErrorWrongDataType, v.toAny())
		}

		return bv, nil
	case SubExpr:
		if res, err = EvalExpression(Expression{&e.BoolExpr}, syms); err != nil {
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

type evalKind uint8

const (
	kindBool    evalKind = iota + 1
	kindFloat64
	kindInt
	kindString
	kindAny
)

// evalVal is a tagged union that avoids boxing primitive values into `any`.
type evalVal struct {
	s    string
	f    float64
	a    any
	i    int
	kind evalKind
	b    bool
}

func (v evalVal) toAny() any {
	switch v.kind {
	case kindBool:
		return v.b
	case kindFloat64:
		return v.f
	case kindInt:
		return v.i
	case kindString:
		return v.s
	default:
		return v.a
	}
}

func (v evalVal) toBool() (bool, bool) {
	switch v.kind {
	case kindBool:
		return v.b, true
	case kindAny:
		b, ok := v.a.(bool)
		return b, ok
	default:
		return false, false
	}
}

func (v evalVal) toFloat() (float64, bool) {
	switch v.kind {
	case kindFloat64:
		return v.f, true
	case kindInt:
		return v.f, true
	case kindAny:
		switch val := v.a.(type) {
		case float64:
			return val, true
		case int:
			return float64(val), true
		}

		return 0, false
	default:
		return 0, false
	}
}

func (v evalVal) toString() (string, bool) {
	switch v.kind {
	case kindString:
		return v.s, true
	case kindAny:
		s, ok := v.a.(string)
		return s, ok
	default:
		return "", false
	}
}

func evalValue(v Value, syms Symbols) (evalVal, error) {
	switch {
	case v.Bool != nil:
		return evalVal{kind: kindBool, b: bool(*v.Bool)}, nil
	case v.Float != nil:
		return evalVal{kind: kindFloat64, f: *v.Float}, nil
	case v.Int != nil:
		return evalVal{kind: kindInt, i: *v.Int, f: float64(*v.Int)}, nil
	case v.String != nil:
		return evalVal{kind: kindString, s: *v.String}, nil
	case v.Symbol != nil:
		val, err := syms.Get(*v.Symbol)
		if err != nil {
			return evalVal{}, err
		}

		return evalVal{kind: kindAny, a: val}, nil
	default:
		return evalVal{}, ErrValueDoesntHaveAnyVal
	}
}

func evalComparisonOpVal(o ComparisonOp, l, r evalVal) (bool, error) {
	if o.Contains {
		return containsEval(l.toAny(), r.toAny())
	} else if o.Excludes {
		res, err := containsEval(l.toAny(), r.toAny())
		return !res, err
	} else if o.StartsWith {
		return startsWithEval(l.toAny(), r.toAny())
	} else if o.EndsWith {
		return endsWithEval(l.toAny(), r.toAny())
	}

	return evalCmpVal(o, l, r)
}

// evalCmpVal handles scalar comparisons on evalVal without boxing literals into any.
// Falls back to the any-based path only for kindAny (resolved symbol values).
func evalCmpVal(o ComparisonOp, l, r evalVal) (bool, error) {
	switch l.kind {
	case kindBool:
		rb, ok := r.toBool()
		if !ok {
			return false, newErrorDataTypeMismatch(opName(o), l.toAny(), r.toAny())
		}

		return applyCmpBool(o, l.b, rb)

	case kindFloat64, kindInt: // both carry f
		rf, ok := r.toFloat()
		if !ok {
			return false, newErrorDataTypeMismatch(opName(o), l.toAny(), r.toAny())
		}

		return applyCmpFloat(o, l.f, rf)

	case kindString:
		rs, ok := r.toString()
		if !ok {
			return false, newErrorDataTypeMismatch(opName(o), l.toAny(), r.toAny())
		}

		return applyCmpStr(o, l.s, rs)

	default: // kindAny: symbol result — fall back to any-based path
		return evalComparisonOp(o, l.toAny(), r.toAny())
	}
}

func applyCmpBool(o ComparisonOp, l, r bool) (bool, error) {
	switch {
	case o.Eq || o.EqEq:
		return l == r, nil
	case o.Neq:
		return l != r, nil
	default:
		return false, newErrorWrongDataType(opName(o), l)
	}
}

func applyCmpFloat(o ComparisonOp, l, r float64) (bool, error) {
	switch {
	case o.Eq || o.EqEq:
		return l == r, nil
	case o.Neq:
		return l != r, nil
	case o.Gt:
		return l > r, nil
	case o.Gte:
		return l >= r, nil
	case o.Lt:
		return l < r, nil
	case o.Lte:
		return l <= r, nil
	default:
		return false, newErrorWrongDataType(opName(o), l)
	}
}

func applyCmpStr(o ComparisonOp, l, r string) (bool, error) {
	switch {
	case o.Eq || o.EqEq:
		return l == r, nil
	case o.Neq:
		return l != r, nil
	case o.Gt:
		return l > r, nil
	case o.Gte:
		return l >= r, nil
	case o.Lt:
		return l < r, nil
	case o.Lte:
		return l <= r, nil
	default:
		return false, newErrorWrongDataType(opName(o), l)
	}
}

func opName(o ComparisonOp) string {
	switch {
	case o.Eq || o.EqEq:
		return "="
	case o.Neq:
		return "!="
	case o.Gt:
		return ">"
	case o.Gte:
		return ">="
	case o.Lt:
		return "<"
	case o.Lte:
		return "<="
	case o.Contains:
		return "contains"
	case o.Excludes:
		return "excludes"
	case o.StartsWith:
		return "starts_with"
	case o.EndsWith:
		return "ends_with"
	default:
		return "?"
	}
}

// evalComparisonOp is the any-based fallback used for kindAny (symbol) values.
func evalComparisonOp(o ComparisonOp, l, r any) (res bool, err error) {
	if o.Eq || o.EqEq {
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
	} else if o.Contains {
		return containsEval(l, r)
	} else if o.Excludes {
		res, err := containsEval(l, r)
		return !res, err
	} else if o.StartsWith {
		return startsWithEval(l, r)
	} else if o.EndsWith {
		return endsWithEval(l, r)
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

func containsEval(l, r any) (bool, error) {
	switch lv := l.(type) {
	case string:
		switch rv := r.(type) {
		case string:
			return strings.Contains(lv, rv), nil
		default:
			return false, newErrorDataTypeMismatch("contains", lv, rv)
		}
	case []string:
		switch rv := r.(type) {
		case string:
			return slices.Contains(lv, rv), nil
		default:
			return false, newErrorDataTypeMismatch("contains", lv, rv)
		}
	case []int:
		switch rv := r.(type) {
		case int:
			return slices.Contains(lv, rv), nil
		case float64:
			for _, v := range lv {
				if float64(v) == rv {
					return true, nil
				}
			}
			return false, nil
		default:
			return false, newErrorDataTypeMismatch("contains", lv, rv)
		}
	case []float64:
		switch rv := r.(type) {
		case float64:
			return slices.Contains(lv, rv), nil
		case int:
			for _, v := range lv {
				if v == float64(rv) {
					return true, nil
				}
			}
			return false, nil
		default:
			return false, newErrorDataTypeMismatch("contains", lv, rv)
		}
	case []bool:
		switch rv := r.(type) {
		case bool:
			return slices.Contains(lv, rv), nil
		default:
			return false, newErrorDataTypeMismatch("contains", lv, rv)
		}
	default:
		return false, newErrorWrongDataType("contains", lv)
	}
}

func startsWithEval(l, r any) (bool, error) {
	lv, ok := l.(string)
	if !ok {
		return false, newErrorWrongDataType("starts_with", l)
	}

	rv, ok := r.(string)
	if !ok {
		return false, newErrorDataTypeMismatch("starts_with", l, r)
	}

	return strings.HasPrefix(lv, rv), nil
}

func endsWithEval(l, r any) (bool, error) {
	lv, ok := l.(string)
	if !ok {
		return false, newErrorWrongDataType("ends_with", l)
	}

	rv, ok := r.(string)
	if !ok {
		return false, newErrorDataTypeMismatch("ends_with", l, r)
	}

	return strings.HasSuffix(lv, rv), nil
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
