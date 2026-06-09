package internal

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

type BoolValue struct {
	Value Value `@@`
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
	Neq      bool `@"!" "="`
	Gte      bool `| @">" "="`
	Lte      bool `| @"<" "="`
	Gt       bool `| @">"`
	Lt       bool `| @"<"`
	EqEq     bool `| @"=" "="`
	Eq       bool `| @"="`
	Contains   bool `| @"contains"`
	Excludes   bool `| @"excludes"`
	StartsWith bool `| @"starts_with"`
	EndsWith   bool `| @"ends_with"`
}

type Boolean bool

func (b *Boolean) Capture(values []string) error {
	*b = values[0] == "true"
	return nil
}
