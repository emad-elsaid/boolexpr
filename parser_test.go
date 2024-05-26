package boolexpr

import (
	"testing"

	"github.com/alecthomas/participle/v2"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	strPtr := func(s string) *string { return &s }
	intPtr := func(i int) *int { return &i }
	boolPtr := func(b Boolean) *Boolean { return &b }

	tcs := []struct {
		name     string
		input    string
		expected *BoolExpr
	}{
		{
			name:  "simple comparison",
			input: "x > 1",
			expected: &BoolExpr{
				Expr: Compare{
					Left: "x",
					Right: &OpValue{
						Op:    Op{Gt: true},
						Value: Value{Int: intPtr(1)},
					},
				},
			},
		},
		{
			name:  "2 comparison with and",
			input: "x > 1 and y = 2",
			expected: &BoolExpr{
				Expr: Compare{
					Left: "x",
					Right: &OpValue{
						Op:    Op{Gt: true},
						Value: Value{Int: intPtr(1)},
					},
				},
				OpExprs: []*OpExpr{
					{
						Op: BoolOp{And: true},
						Expr: Compare{
							Left: "y",
							Right: &OpValue{
								Op:    Op{Eq: true},
								Value: Value{Int: intPtr(2)},
							},
						},
					},
				},
			},
		},
		{
			name:  "2 comparison with group",
			input: `x > 1 and y = 2 or ( x = "hello" or z = true ) and test = false`,
			expected: &BoolExpr{
				Expr: Compare{
					Left: "x",
					Right: &OpValue{
						Op:    Op{Gt: true},
						Value: Value{Int: intPtr(1)},
					},
				},
				OpExprs: []*OpExpr{
					{
						Op: BoolOp{And: true},
						Expr: Compare{
							Left: "y",
							Right: &OpValue{
								Op:    Op{Eq: true},
								Value: Value{Int: intPtr(2)},
							},
						},
					},
					{
						Op: BoolOp{Or: true},
						Expr: Group{
							BoolExpr: &BoolExpr{
								Expr: Compare{
									Left: "x",
									Right: &OpValue{
										Op:    Op{Eq: true},
										Value: Value{String: strPtr("hello")},
									},
								},
								OpExprs: []*OpExpr{
									{
										Op: BoolOp{Or: true},
										Expr: Compare{
											Left: "z",
											Right: &OpValue{
												Op: Op{Eq: true},
												Value: Value{
													Bool: boolPtr(true),
												},
											},
										},
									},
								},
							},
						},
					},
					{
						Op: BoolOp{And: true},
						Expr: Compare{
							Left: "test",
							Right: &OpValue{
								Op:    Op{Eq: true},
								Value: Value{Bool: boolPtr(false)},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		parser, err := participle.Build[BoolExpr](
			participle.Unquote("String"),
			participle.Union[Expr](Compare{}, Group{}),
		)
		assert.NoError(t, err)

		output, err := parser.ParseString("", tc.input)
		assert.NoError(t, err)

		assert.Equal(t, tc.expected, output)
	}
}
