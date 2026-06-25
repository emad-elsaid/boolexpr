package boolexpr

import (
	"testing"

	. "github.com/emad-elsaid/boolexpr/internal"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	strPtr := func(s string) *string { return &s }
	intPtr := func(i int) *int { return &i }
	boolPtr := func(b Boolean) *Boolean { return &b }

	// and builds an AndExpr from a leading expression followed by AND-ed ones.
	and := func(first Expr, rest ...Expr) AndExpr {
		a := AndExpr{Expr: first}
		for _, e := range rest {
			a.AndOps = append(a.AndOps, AndOpExpr{Expr: e})
		}
		return a
	}

	// expr builds a BoolExpr from a leading AND-expression OR-ed with more.
	expr := func(first AndExpr, rest ...AndExpr) *BoolExpr {
		b := &BoolExpr{And: first}
		for _, a := range rest {
			b.OrOps = append(b.OrOps, OrOpExpr{And: a})
		}
		return b
	}

	tcs := []struct {
		name     string
		input    string
		expected *BoolExpr
	}{
		{
			name:  "simple comparison",
			input: "x > 1",
			expected: expr(and(Compare{
				Left:  Value{Symbol: strPtr("x")},
				Op:    ComparisonOp{Gt: true},
				Right: Value{Int: intPtr(1)},
			})),
		},
		{
			name:  "simple comparison with !=",
			input: "x != 1",
			expected: expr(and(Compare{
				Left:  Value{Symbol: strPtr("x")},
				Op:    ComparisonOp{Neq: true},
				Right: Value{Int: intPtr(1)},
			})),
		},
		{
			name:  "simple comparison with >=",
			input: "x >= 1",
			expected: expr(and(Compare{
				Left:  Value{Symbol: strPtr("x")},
				Op:    ComparisonOp{Gte: true},
				Right: Value{Int: intPtr(1)},
			})),
		},
		{
			name:  "simple comparison with two variables",
			input: "x > y",
			expected: expr(and(Compare{
				Left:  Value{Symbol: strPtr("x")},
				Op:    ComparisonOp{Gt: true},
				Right: Value{Symbol: strPtr("y")},
			})),
		},
		{
			name:  "2 comparison with and",
			input: "x > 1 and y = 2",
			expected: expr(and(
				Compare{
					Left:  Value{Symbol: strPtr("x")},
					Op:    ComparisonOp{Gt: true},
					Right: Value{Int: intPtr(1)},
				},
				Compare{
					Left:  Value{Symbol: strPtr("y")},
					Op:    ComparisonOp{Eq: true},
					Right: Value{Int: intPtr(2)},
				},
			)),
		},
		{
			// "and" binds tighter than "or": (x > 1 && y = 2) || z = 3
			name:  "2 comparison with && and ||",
			input: "x > 1 && y = 2 || z = 3",
			expected: expr(
				and(
					Compare{
						Left:  Value{Symbol: strPtr("x")},
						Op:    ComparisonOp{Gt: true},
						Right: Value{Int: intPtr(1)},
					},
					Compare{
						Left:  Value{Symbol: strPtr("y")},
						Op:    ComparisonOp{Eq: true},
						Right: Value{Int: intPtr(2)},
					},
				),
				and(Compare{
					Left:  Value{Symbol: strPtr("z")},
					Op:    ComparisonOp{Eq: true},
					Right: Value{Int: intPtr(3)},
				}),
			),
		},
		{
			// "or" splits at the top level; each side is an AND-expression:
			//   (x > 1 and y = 2)  or  (( x = "hello" or z = true ) and test = false)
			name:  "2 comparison with group",
			input: `x > 1 and y = 2 or ( x = "hello" or z = true ) and test = false`,
			expected: expr(
				and(
					Compare{
						Left:  Value{Symbol: strPtr("x")},
						Op:    ComparisonOp{Gt: true},
						Right: Value{Int: intPtr(1)},
					},
					Compare{
						Left:  Value{Symbol: strPtr("y")},
						Op:    ComparisonOp{Eq: true},
						Right: Value{Int: intPtr(2)},
					},
				),
				and(
					SubExpr{
						BoolExpr: *expr(
							and(Compare{
								Left:  Value{Symbol: strPtr("x")},
								Op:    ComparisonOp{Eq: true},
								Right: Value{String: strPtr("hello")},
							}),
							and(Compare{
								Left:  Value{Symbol: strPtr("z")},
								Op:    ComparisonOp{Eq: true},
								Right: Value{Bool: boolPtr(true)},
							}),
						),
					},
					Compare{
						Left:  Value{Symbol: strPtr("test")},
						Op:    ComparisonOp{Eq: true},
						Right: Value{Bool: boolPtr(false)},
					},
				),
			),
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			output, err := Parse(tc.input)
			assert.NoError(t, err)
			assert.Equal(t, Expression{tc.expected}, output)
		})
	}
}
