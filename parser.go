package boolexpr

import "github.com/alecthomas/participle/v2"

// Parse will convert string s to BoolExpr tree that can be evaluated multiple times
func Parse(s string) (*BoolExpr, error) {
	return parser.ParseString("", s)
}

var parser = must(participle.Build[BoolExpr](
	participle.Unquote("String"),
	participle.Union[Expr](Compare{}, Group{}),
))

func must[T any](p T, err error) T {
	if err != nil {
		panic(err)
	}

	return p
}

type BoolExpr struct {
	Expr    Expr      `@@`
	OpExprs []*OpExpr `@@*`
}

type Expr interface {
	Eval(Symbols) (bool, error)
}

type Compare struct {
	Left  string   `@Ident`
	Right *OpValue `@@`
}

type Group struct {
	BoolExpr *BoolExpr `"(" @@ ")"`
}

type OpExpr struct {
	Op   BoolOp `@@`
	Expr Expr   `@@`
}

type BoolOp struct {
	And bool `@"and"`
	Or  bool `| @"or"`
}

type Boolean bool

func (b *Boolean) Capture(values []string) error {
	*b = values[0] == "true"
	return nil
}

type Value struct {
	Float  *float64 `  @Float`
	Int    *int     `| @Int`
	String *string  `| @String`
	Bool   *Boolean `| @("true" | "false")`
}

type OpValue struct {
	Op    Op    `@@`
	Value Value `@@`
}

type Op struct {
	Neq bool `@"!" "="`
	Gte bool `| @">" "="`
	Lte bool `| @"<" "="`
	Gt  bool `| @">"`
	Lt  bool `| @"<"`
	Eq  bool `| @"="`
}
