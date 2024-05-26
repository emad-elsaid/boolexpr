package boolexpr

import (
	"fmt"
	"testing"

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
		{
			input:    `x = "Hello"`,
			expected: true,
			symbols: map[string]func() any{
				"x": func() any { return "Hello" },
			},
		},
		{
			input:    `x = "Hello"`,
			expected: false,
			symbols: map[string]func() any{
				"x": func() any { return "World" },
			},
		},
		{
			input:    `x = 2`,
			expected: false,
			symbols: map[string]func() any{
				"x": func() any { return 2.4 },
			},
		},
		{
			input:    `x = 2`,
			expected: true,
			symbols: map[string]func() any{
				"x": func() any { return 2.0 },
			},
		},
		{
			input:    `x >= 10`,
			expected: true,
			symbols: map[string]func() any{
				"x": func() any { return 11 },
			},
		},
		{
			input:    `x != 10`,
			expected: true,
			symbols: map[string]func() any{
				"x": func() any { return 11 },
			},
		},
		{
			input:    `x <= 10`,
			expected: false,
			symbols: map[string]func() any{
				"x": func() any { return 11 },
			},
		},
		{
			input:    `x >= 10 and y < 0`,
			expected: true,
			symbols: map[string]func() any{
				"x": func() any { return 11 },
				"y": func() any { return -1 },
			},
		},
		{
			input:    `x >= 10 and y < 0`,
			expected: false,
			symbols: map[string]func() any{
				"x": func() any { return 11 },
				"y": func() any { return 0 },
			},
		},
		{
			input:    `x >= 10 or y < 0`,
			expected: true,
			symbols: map[string]func() any{
				"x": func() any { return 11 },
				"y": func() any { return 0 },
			},
		},
		{
			input:    `x >= 10 or y < 0 or ( z = "hello" or z = "world" )`,
			expected: true,
			symbols: map[string]func() any{
				"x": func() any { return 0 },
				"y": func() any { return 0 },
				"z": func() any { return "hello" },
			},
		},
		{
			input:    `x >= 10 or y < 0 or ( z = "hello" or z = "world" )`,
			expected: false,
			symbols: map[string]func() any{
				"x": func() any { return 0 },
				"y": func() any { return 0 },
				"z": func() any { return "NO" },
			},
		},
		{
			input:    `x > y and y > z`,
			expected: true,
			symbols: map[string]func() any{
				"x": func() any { return 10 },
				"y": func() any { return 5 },
				"z": func() any { return 2 },
			},
		},
		{
			input:    `x > y and y > z`,
			expected: false,
			symbols: map[string]func() any{
				"x": func() any { return 10 },
				"y": func() any { return 5 },
				"z": func() any { return 6 },
			},
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(fmt.Sprintf("%s -> %t", tc.input, tc.expected), func(t *testing.T) {
			output, err := Eval(tc.input, tc.symbols)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, output)
		})
	}
}
