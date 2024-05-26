package boolexpr

type BoolExpr struct {
	Expr    Expr      `@@`
	OpExprs []*OpExpr `@@*`
}

type Expr interface{}

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
	And *string `@"and"`
	Or  *string `| @"or"`
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
	Neq *string `@"!="`
	Eq  *string `| @"="`
	Gte *string `| @">="`
	Gt  *string `| @">"`
	Lte *string `| @"<="`
	Lt  *string `| @"<"`
}
