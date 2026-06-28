package boolexpr

import (
	"cmp"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	. "github.com/emad-elsaid/boolexpr/internal"
)

// Errors returned during evaluation. Use errors.Is to test for them; the
// returned errors are wrapped with additional context (operator, value types).
var (
	// ErrSymbolNotFound is returned when an expression references a symbol
	// that is absent from the provided Symbols.
	ErrSymbolNotFound = errors.New("Symbol not found")
	// ErrLogicalOperationUndefinedState indicates a logical operator node that
	// is neither "and" nor "or"; it signals a malformed expression tree.
	ErrLogicalOperationUndefinedState = errors.New("Logical operation not specified")
	// ErrValueDoesntHaveAnyVal indicates a value node that carries no literal
	// or symbol; it signals a malformed expression tree.
	ErrValueDoesntHaveAnyVal = errors.New("Value is not specified")
	// ErrSymbolTypeUnknown is returned for an unsupported symbol value type.
	ErrSymbolTypeUnknown = errors.New("Symbol type not supported")
	// ErrOpDoesnotHaveVal indicates a comparison node with no operator set; it
	// signals a malformed expression tree.
	ErrOpDoesnotHaveVal = errors.New("Operation not specified")
	// ErrorWrongDataType is returned when an operator is applied to operands of
	// incompatible types, e.g. comparing a string with an int.
	ErrorWrongDataType = errors.New("Wrong data type")
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

// Eval parses the expression string s and evaluates it against syms in a
// single call. It is a convenience wrapper around [Parse] followed by
// [EvalExpression]. When the same expression is evaluated more than once,
// prefer parsing it once with [Parse] and reusing the result.
func Eval(s string, syms Symbols) (bool, error) {
	ast, err := Parse(s)
	if err != nil {
		return false, err
	}

	return EvalExpression(ast, syms)
}

// EvalExpression evaluates an already-parsed [Expression] against syms and
// returns the boolean result. A single parsed Expression may be evaluated
// concurrently against different Symbols.
func EvalExpression(e Expression, syms Symbols) (bool, error) {
	if e.e == nil {
		return false, errors.New("EvalExpression called on zero-value Expression; use Parse to obtain a valid Expression")
	}
	return evalBoolExpr(e.e, syms)
}

// evalBoolExpr evaluates the OR level: the leading AND-expression OR-ed with the
// rest. It short-circuits as soon as one AND-expression is true.
func evalBoolExpr(b *BoolExpr, syms Symbols) (bool, error) {
	res, err := evalAndExpr(b.And, syms)
	if err != nil {
		return false, err
	}

	for _, o := range b.OrOps {
		if res {
			// short circuit: the whole OR is already true
			return true, nil
		}

		res, err = evalAndExpr(o.And, syms)
		if err != nil {
			return false, err
		}
	}

	return res, nil
}

// evalAndExpr evaluates the AND level: the leading primary expression AND-ed
// with the rest. It short-circuits as soon as one operand is false.
func evalAndExpr(a AndExpr, syms Symbols) (bool, error) {
	res, err := evalExpr(a.Expr, syms)
	if err != nil {
		return false, err
	}

	for _, op := range a.AndOps {
		if !res {
			// short circuit: the whole AND is already false
			return false, nil
		}

		res, err = evalExpr(op.Expr, syms)
		if err != nil {
			return false, err
		}
	}

	return res, nil
}

func evalExpr(b Expr, syms Symbols) (bool, error) {
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
		return evalBoolExpr(&e.BoolExpr, syms)
	default:
		return false, fmt.Errorf("Expr type is unhandled %T", b)
	}
}

type evalKind uint8

const (
	kindBool evalKind = iota + 1
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

// toInt returns the value as an int when the operand is an integer — a literal
// int or a symbol resolving to int. Floats report false: they are only ever
// compared through toFloat. Used to keep integer comparisons exact rather than
// routing large integers through float64, which loses precision above 2^53.
func (v evalVal) toInt() (int, bool) {
	switch v.kind {
	case kindInt:
		return v.i, true
	case kindAny:
		i, ok := v.a.(int)
		return i, ok
	default:
		return 0, false
	}
}

func (v evalVal) toFloat() (float64, bool) {
	switch v.kind {
	case kindFloat64:
		return v.f, true
	case kindInt:
		return float64(v.i), true
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
		return evalVal{kind: kindInt, i: *v.Int}, nil
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
		return containsEval(l, r)
	} else if o.Excludes {
		res, err := containsEval(l, r)
		if err != nil {
			return false, err
		}
		return !res, nil
	} else if o.StartsWith {
		return startsWithEval(l, r)
	} else if o.EndsWith {
		return endsWithEval(l, r)
	} else if o.Match {
		return matchEval(l, r)
	}

	return evalCmpVal(o, l, r)
}

// evalCmpVal handles scalar comparisons on evalVal without boxing operands into
// any. For literal operands the concrete type is known from the kind; for
// kindAny (resolved symbol values) it type-switches on the stored value, while
// still reading the right operand through the non-boxing accessors.
func evalCmpVal(o ComparisonOp, l, r evalVal) (bool, error) {
	switch l.kind {
	case kindBool:
		return cmpBoolEval(o, l.b, r)

	case kindFloat64, kindInt:
		return cmpNumEval(o, l, r)

	case kindString:
		return cmpStrEval(o, l.s, r)

	default: // kindAny: resolved symbol value
		switch lv := l.a.(type) {
		case bool:
			return cmpBoolEval(o, lv, r)
		case int, float64:
			return cmpNumEval(o, l, r)
		case string:
			return cmpStrEval(o, lv, r)
		default:
			return false, newErrorWrongDataType(opName(o), l.toAny())
		}
	}
}

// cmpBoolEval compares a bool left operand against r without boxing r.
func cmpBoolEval(o ComparisonOp, l bool, r evalVal) (bool, error) {
	rb, ok := r.toBool()
	if !ok {
		return false, newErrorDataTypeMismatch(opName(o), l, r.toAny())
	}

	return applyCmpBool(o, l, rb)
}

// cmpNumEval compares two numeric operands without boxing. When both operands
// are integers it compares them exactly as int; otherwise it falls back to
// float64, which also handles mixed int/float comparisons. The int lane avoids
// the precision loss that float64 incurs for integers beyond 2^53.
func cmpNumEval(o ComparisonOp, l, r evalVal) (bool, error) {
	if li, ok := l.toInt(); ok {
		if ri, ok := r.toInt(); ok {
			return applyCmpOrdered(o, li, ri)
		}
	}

	lf, ok := l.toFloat()
	if !ok {
		return false, newErrorWrongDataType(opName(o), l.toAny())
	}

	rf, ok := r.toFloat()
	if !ok {
		return false, newErrorDataTypeMismatch(opName(o), l.toAny(), r.toAny())
	}

	return applyCmpOrdered(o, lf, rf)
}

// cmpStrEval compares a string left operand against r without boxing r.
func cmpStrEval(o ComparisonOp, l string, r evalVal) (bool, error) {
	rs, ok := r.toString()
	if !ok {
		return false, newErrorDataTypeMismatch(opName(o), l, r.toAny())
	}

	return applyCmpOrdered(o, l, rs)
}

func applyCmpBool(o ComparisonOp, l, r bool) (bool, error) {
	switch {
	case o.Eq || o.EqEq:
		return l == r, nil
	case o.Neq:
		return l != r, nil
	default:
		return false, fmt.Errorf("%w, %s is not defined for bool", ErrOpDoesnotHaveVal, opName(o))
	}
}

func applyCmpOrdered[T cmp.Ordered](o ComparisonOp, l, r T) (bool, error) {
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
		return false, fmt.Errorf("%w, %s is not defined for %T", ErrOpDoesnotHaveVal, opName(o), l)
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
	case o.Match:
		return "match"
	default:
		return "?"
	}
}

// containsEval reports whether l contains r. The left operand may be a string
// (substring match) or a slice resolved from a symbol (element match); the
// right operand is read through the non-boxing accessors so literals are not
// boxed into any.
func containsEval(l, r evalVal) (bool, error) {
	// String containment: left is a literal string or a string symbol.
	if ls, ok := l.toString(); ok {
		rs, ok := r.toString()
		if !ok {
			return false, newErrorDataTypeMismatch("contains", l.toAny(), r.toAny())
		}

		return strings.Contains(ls, rs), nil
	}

	// Slice containment: slices only ever arrive as resolved symbol values.
	switch lv := l.a.(type) {
	case []string:
		rs, ok := r.toString()
		if !ok {
			return false, newErrorDataTypeMismatch("contains", lv, r.toAny())
		}

		return slices.Contains(lv, rs), nil
	case []int:
		// Compare exactly when the right operand is an integer; only fall back
		// to float64 for an actual float operand (e.g. `ids contains 1.0`).
		if ri, ok := r.toInt(); ok {
			return slices.Contains(lv, ri), nil
		}

		rf, ok := r.toFloat()
		if !ok {
			return false, newErrorDataTypeMismatch("contains", lv, r.toAny())
		}

		return slices.ContainsFunc(lv, func(v int) bool { return float64(v) == rf }), nil
	case []float64:
		rf, ok := r.toFloat()
		if !ok {
			return false, newErrorDataTypeMismatch("contains", lv, r.toAny())
		}

		return slices.Contains(lv, rf), nil
	case []bool:
		rb, ok := r.toBool()
		if !ok {
			return false, newErrorDataTypeMismatch("contains", lv, r.toAny())
		}

		return slices.Contains(lv, rb), nil
	default:
		return false, newErrorWrongDataType("contains", l.toAny())
	}
}

func startsWithEval(l, r evalVal) (bool, error) {
	lv, rv, err := stringOperands("starts_with", l, r)
	if err != nil {
		return false, err
	}

	return strings.HasPrefix(lv, rv), nil
}

func endsWithEval(l, r evalVal) (bool, error) {
	lv, rv, err := stringOperands("ends_with", l, r)
	if err != nil {
		return false, err
	}

	return strings.HasSuffix(lv, rv), nil
}

// matchEval reports whether the left string matches the regular expression in
// the right operand. Both operands must be strings; the pattern uses Go's
// regexp (RE2) syntax. The match is unanchored, like [regexp.Regexp.MatchString].
// Compiled patterns are cached so each distinct pattern is compiled only once.
func matchEval(l, r evalVal) (bool, error) {
	lv, rv, err := stringOperands("match", l, r)
	if err != nil {
		return false, err
	}

	re, err := compilePattern(rv)
	if err != nil {
		return false, fmt.Errorf("%w, invalid match pattern %q: %v", ErrorWrongDataType, rv, err)
	}

	return re.MatchString(lv), nil
}

// stringOperands extracts two string operands without boxing, returning a typed
// error for the string-only operators (starts_with, ends_with, match).
func stringOperands(op string, l, r evalVal) (string, string, error) {
	lv, ok := l.toString()
	if !ok {
		return "", "", newErrorWrongDataType(op, l.toAny())
	}

	rv, ok := r.toString()
	if !ok {
		return "", "", newErrorDataTypeMismatch(op, l.toAny(), r.toAny())
	}

	return lv, rv, nil
}

// matchCache memoizes compiled patterns so a regular expression is compiled
// once per distinct pattern, not on every evaluation. Patterns may originate
// from symbols, so the pattern string is only known at evaluation time; keying
// the cache on it covers both literal and symbol-supplied patterns.
// The cache is capped at matchCacheMax entries to bound memory growth when
// patterns come from high-cardinality symbol values.
var (
	matchCache    sync.Map // map[string]*regexp.Regexp
	matchCacheLen atomic.Int64
)

const matchCacheMax = 1024

func compilePattern(pattern string) (*regexp.Regexp, error) {
	if re, ok := matchCache.Load(pattern); ok {
		return re.(*regexp.Regexp), nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	if matchCacheLen.Load() < matchCacheMax {
		if _, loaded := matchCache.LoadOrStore(pattern, re); !loaded {
			matchCacheLen.Add(1)
		}
	}
	return re, nil
}
