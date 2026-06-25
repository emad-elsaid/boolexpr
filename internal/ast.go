package internal

// BoolExpr is the grammar root. "or" (and "||") has the lowest precedence, so
// an expression is a sequence of AND-expressions joined by "or", matching the
// precedence used by Go and most languages where "and" binds tighter than "or".
type BoolExpr struct {
	And   AndExpr    `parser:"@@"`
	OrOps []OrOpExpr `parser:"@@*"`
}

// OrOpExpr is an "or"/"||" operator followed by its right-hand AND-expression.
type OrOpExpr struct {
	And AndExpr `parser:"('or' | '|' '|') @@"`
}

// AndExpr binds "and"/"&&" tighter than "or": a sequence of primary
// expressions joined by "and".
type AndExpr struct {
	Expr   Expr        `parser:"@@"`
	AndOps []AndOpExpr `parser:"@@*"`
}

// AndOpExpr is an "and"/"&&" operator followed by its right-hand expression.
type AndOpExpr struct {
	Expr Expr `parser:"('and' | '&' '&') @@"`
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
