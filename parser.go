package boolexpr

import "github.com/alecthomas/participle/v2"

// Parse will convert string s to BoolExpr tree that can be evaluated multiple times
func Parse(s string) (*BoolExpr, error) {
	return parser.ParseString("", s)
}

var parser, parserErr = participle.Build[BoolExpr](
	participle.Unquote("String"),
	participle.Union[Expr](Compare{}, SubExpr{}),
)

type BoolExpr struct {
	Expr    Expr     `@@`
	OpExprs []OpExpr `@@*`
}

type Expr interface{}

type SubExpr struct {
	BoolExpr BoolExpr `"(" @@ ")"`
}

type Compare struct {
	Left  Value        `@@`
	Op    ComparisonOp `@@`
	Right Value        `@@`
}

type OpExpr struct {
	Op   LogicalOp `@@`
	Expr Expr      `@@`
}

type LogicalOp struct {
	And bool `@"and"`
	Or  bool `| @"or"`
}

type Value struct {
	Float  *float64 `  @Float`
	Int    *int     `| @Int`
	String *string  `| @String`
	Bool   *Boolean `| @("true" | "false")`
	Symbol *string  `| @Ident`
}

type ComparisonOp struct {
	Neq bool `@"!" "="`
	Gte bool `| @">" "="`
	Lte bool `| @"<" "="`
	Gt  bool `| @">"`
	Lt  bool `| @"<"`
	Eq  bool `| @"="`
}

type Boolean bool

func (b *Boolean) Capture(values []string) error {
	*b = values[0] == "true"
	return nil
}
