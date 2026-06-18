package internal

type BoolExpr struct {
	Expr    Expr     `parser:"@@"`
	OpExprs []OpExpr `parser:"@@*"`
}

type Expr interface{}

type SubExpr struct {
	BoolExpr BoolExpr `parser:"'(' @@ ')'"`
}

type Compare struct {
	Left  Value        `parser:"@@"`
	Op    ComparisonOp `parser:"@@"`
	Right Value        `parser:"@@"`
}

type BoolValue struct {
	Value Value `parser:"@@"`
}

type OpExpr struct {
	Op   LogicalOp `parser:"@@"`
	Expr Expr      `parser:"@@"`
}

type LogicalOp struct {
	And bool `parser:"@'and' | @'&' '&'"`
	Or  bool `parser:"| @'or' | @'|' '|'"`
}

type Value struct {
	Float  *float64 `parser:"  @Float"`
	Int    *int     `parser:"| @Int"`
	String *string  `parser:"| @String"`
	Bool   *Boolean `parser:"| @('true' | 'false')"`
	Symbol *string  `parser:"| @Ident"`
}

type ComparisonOp struct {
	Neq        bool `parser:"@'!' '='"`
	Gte        bool `parser:"| @'>' '='"`
	Lte        bool `parser:"| @'<' '='"`
	Gt         bool `parser:"| @'>'"`
	Lt         bool `parser:"| @'<'"`
	EqEq       bool `parser:"| @'=' '='"`
	Eq         bool `parser:"| @'='"`
	Contains   bool `parser:"| @'contains'"`
	Excludes   bool `parser:"| @'excludes'"`
	StartsWith bool `parser:"| @'starts_with'"`
	EndsWith   bool `parser:"| @'ends_with'"`
	Match      bool `parser:"| @'match'"`
}

type Boolean bool

func (b *Boolean) Capture(values []string) error {
	*b = values[0] == "true"
	return nil
}
