package boolexpr

import (
	"testing"

	"github.com/alecthomas/participle/v2"
	"github.com/stretchr/testify/assert"
)

func TestEval(t *testing.T) {
	tcs := []struct {
		name     string
		input    string
		symbols  map[string]func() any
		expected bool
	}{
		{
			name:  "x = 1 -> true",
			input: "x = 1",
			symbols: map[string]func() any{
				"x": func() any { return 1 },
			},
			expected: true,
		},
		{
			name:  "x = 2 -> false",
			input: "x = 2",
			symbols: map[string]func() any{
				"x": func() any { return 1 },
			},
			expected: false,
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			parser, err := participle.Build[BoolExpr](
				participle.Unquote("String"),
				participle.Union[Expr](Compare{}, Group{}),
			)
			assert.NoError(t, err)

			ast, err := parser.ParseString("", tc.input)
			assert.NoError(t, err)

			output, err := ast.Eval(tc.symbols)
			assert.NoError(t, err)

			assert.Equal(t, tc.expected, output)
		})
	}
}
