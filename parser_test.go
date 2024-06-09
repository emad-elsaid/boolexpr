package boolexpr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParserErr(t *testing.T) {
	assert.NoError(t, parserErr)
}

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
					Left:  Value{Symbol: strPtr("x")},
					Op:    ComparisonOp{Gt: true},
					Right: Value{Int: intPtr(1)},
				},
			},
		},
		{
			name:  "simple comparison with !=",
			input: "x != 1",
			expected: &BoolExpr{
				Expr: Compare{
					Left:  Value{Symbol: strPtr("x")},
					Op:    ComparisonOp{Neq: true},
					Right: Value{Int: intPtr(1)},
				},
			},
		},
		{
			name:  "simple comparison with >=",
			input: "x >= 1",
			expected: &BoolExpr{
				Expr: Compare{
					Left:  Value{Symbol: strPtr("x")},
					Op:    ComparisonOp{Gte: true},
					Right: Value{Int: intPtr(1)},
				},
			},
		},
		{
			name:  "simple comparison with two variables",
			input: "x > y",
			expected: &BoolExpr{
				Expr: Compare{
					Left:  Value{Symbol: strPtr("x")},
					Op:    ComparisonOp{Gt: true},
					Right: Value{Symbol: strPtr("y")},
				},
			},
		},
		{
			name:  "2 comparison with and",
			input: "x > 1 and y = 2",
			expected: &BoolExpr{
				Expr: Compare{
					Left:  Value{Symbol: strPtr("x")},
					Op:    ComparisonOp{Gt: true},
					Right: Value{Int: intPtr(1)},
				},
				OpExprs: []OpExpr{
					{
						Op: LogicalOp{And: true},
						Expr: Compare{
							Left:  Value{Symbol: strPtr("y")},
							Op:    ComparisonOp{Eq: true},
							Right: Value{Int: intPtr(2)},
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
					Left:  Value{Symbol: strPtr("x")},
					Op:    ComparisonOp{Gt: true},
					Right: Value{Int: intPtr(1)},
				},
				OpExprs: []OpExpr{
					{
						Op: LogicalOp{And: true},
						Expr: Compare{
							Left:  Value{Symbol: strPtr("y")},
							Op:    ComparisonOp{Eq: true},
							Right: Value{Int: intPtr(2)},
						},
					},
					{
						Op: LogicalOp{Or: true},
						Expr: SubExpr{
							BoolExpr: BoolExpr{
								Expr: Compare{
									Left:  Value{Symbol: strPtr("x")},
									Op:    ComparisonOp{Eq: true},
									Right: Value{String: strPtr("hello")},
								},
								OpExprs: []OpExpr{
									{
										Op: LogicalOp{Or: true},
										Expr: Compare{
											Left:  Value{Symbol: strPtr("z")},
											Op:    ComparisonOp{Eq: true},
											Right: Value{Bool: boolPtr(true)},
										},
									},
								},
							},
						},
					},
					{
						Op: LogicalOp{And: true},
						Expr: Compare{
							Left:  Value{Symbol: strPtr("test")},
							Op:    ComparisonOp{Eq: true},
							Right: Value{Bool: boolPtr(false)},
						},
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			output, err := Parse(tc.input)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, output)
		})
	}
}
