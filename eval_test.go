package boolexpr

import (
	"fmt"
	"testing"

	"github.com/alecthomas/participle/v2"
	"github.com/stretchr/testify/assert"
)

func TestEval(t *testing.T) {
	tcs := []struct {
		input    string
		expected bool
		symbols  map[string]func() any
	}{
		{
			input:    "x = 1",
			expected: true,
			symbols: map[string]func() any{
				"x": func() any { return 1 },
			},
		},
		{
			input:    "x = 2",
			expected: false,
			symbols: map[string]func() any{
				"x": func() any { return 1 },
			},
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(fmt.Sprintf("%s -> %t", tc.input, tc.expected), func(t *testing.T) {
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
