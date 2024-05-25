package boolexpr

import (
	"testing"

	parsec "github.com/prataprc/goparsec"
	"github.com/stretchr/testify/assert"
)

func TestLiteral(t *testing.T) {
	tcs := []struct {
		name     string
		input    string
		expected any
	}{
		{
			name:     "one char symbol",
			input:    "x",
			expected: Sym("x"),
		},
		{
			name:     "multi char symbol",
			input:    "version",
			expected: Sym("version"),
		},
		{
			name:     "symbol with _",
			input:    "user_name",
			expected: Sym("user_name"),
		},
		{
			name:     "symbol with -",
			input:    "user-name",
			expected: Sym("user-name"),
		},
		{
			name:     "symbol with caps",
			input:    "UserName",
			expected: Sym("UserName"),
		},
		{
			name:     "Int",
			input:    "123",
			expected: 123.0,
		},
		{
			name:     "Float",
			input:    "123.123",
			expected: 123.123,
		},
		{
			name:     "String",
			input:    `"hello world"`,
			expected: "hello world",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			output, _ := literal(parsec.NewScanner([]byte(tc.input)))
			assert.Equal(t, tc.expected, output)
		})
	}
}

func TestComparison(t *testing.T) {
	tcs := []struct {
		name     string
		input    string
		expected any
	}{
		{
			name:  "=",
			input: "days = 10",
			expected: Comparison{
				Left:  Sym("days"),
				Right: 10.0,
				Op:    "=",
			},
		},
		{
			name:  ">",
			input: "days > 10",
			expected: Comparison{
				Left:  Sym("days"),
				Right: 10.0,
				Op:    ">",
			},
		},
		{
			name:  "<",
			input: "days < 10",
			expected: Comparison{
				Left:  Sym("days"),
				Right: 10.0,
				Op:    "<",
			},
		},
		{
			name:  ">=",
			input: "days >= 10",
			expected: Comparison{
				Left:  Sym("days"),
				Right: 10.0,
				Op:    ">=",
			},
		},
		{
			name:  "<=",
			input: "days <= 10",
			expected: Comparison{
				Left:  Sym("days"),
				Right: 10.0,
				Op:    "<=",
			},
		},
		{
			name:  "!=",
			input: "days != 10",
			expected: Comparison{
				Left:  Sym("days"),
				Right: 10.0,
				Op:    "!=",
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			output, _ := comparison(parsec.NewScanner([]byte(tc.input)))
			assert.Equal(t, tc.expected, output)
		})
	}
}

func TestLogicalExprs(t *testing.T) {
	tcs := []struct {
		name     string
		input    string
		expected any
	}{
		{
			name:  "AND",
			input: "x > 2 and y < 10",
			expected: LogicalExpr{
				Comparison{
					Left:  Sym("x"),
					Right: 2.0,
					Op:    ">",
				},
				"and",
				Comparison{
					Left:  Sym("y"),
					Right: 10.0,
					Op:    "<",
				},
			},
		},
		{
			name:  "OR",
			input: "x > 2 or y < 10",
			expected: LogicalExpr{
				Comparison{
					Left:  Sym("x"),
					Right: 2.0,
					Op:    ">",
				},
				"or",
				Comparison{
					Left:  Sym("y"),
					Right: 10.0,
					Op:    "<",
				},
			},
		},
		{
			name:  "AND OR AND OR",
			input: "x > 2 and y < 10 or x = 1 and z != 10 or y = 2",
			expected: LogicalExpr{
				Comparison{
					Left:  Sym("x"),
					Right: 2.0,
					Op:    ">",
				},
				"and",
				Comparison{
					Left:  Sym("y"),
					Right: 10.0,
					Op:    "<",
				},
				"or",
				Comparison{
					Left:  Sym("x"),
					Right: 1.0,
					Op:    "=",
				},
				"and",
				Comparison{
					Left:  Sym("z"),
					Right: 10.0,
					Op:    "!=",
				},
				"or",
				Comparison{
					Left:  Sym("y"),
					Right: 2.0,
					Op:    "=",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			output, _ := logicalExprs(parsec.NewScanner([]byte(tc.input)))
			assert.Equal(t, tc.expected, output)
		})
	}
}

func TestGroup(t *testing.T) {
	tcs := []struct {
		name     string
		input    string
		expected any
	}{
		{
			name:  "AND OR AND OR in a group",
			input: "( x > 2 and y < 10 or x = 1 and z != 10 or y = 2 )",
			expected: Group{
				Children: LogicalExpr{
					Comparison{
						Left:  Sym("x"),
						Right: 2.0,
						Op:    ">",
					},
					"and",
					Comparison{
						Left:  Sym("y"),
						Right: 10.0,
						Op:    "<",
					},
					"or",
					Comparison{
						Left:  Sym("x"),
						Right: 1.0,
						Op:    "=",
					},
					"and",
					Comparison{
						Left:  Sym("z"),
						Right: 10.0,
						Op:    "!=",
					},
					"or",
					Comparison{
						Left:  Sym("y"),
						Right: 2.0,
						Op:    "=",
					},
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			output, _ := group(parsec.NewScanner([]byte(tc.input)))
			assert.Equal(t, tc.expected, output)
		})
	}
}

func TestParser(t *testing.T) {
	tcs := []struct {
		name     string
		input    string
		expected any
	}{
		{
			name:  "combination of expression with group",
			input: `x > 2 and (y < 10 or z = "yes") or a = true`,
			expected: Exp{
				Comparison{
					Left:  Sym("x"),
					Right: 2.0,
					Op:    ">",
				},
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			output, _ := Parser(parsec.NewScanner([]byte(tc.input)))
			assert.Equal(t, tc.expected, output)
		})
	}
}
