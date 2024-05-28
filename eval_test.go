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
			input:    "x = 1.0",
			expected: true,
			symbols: map[string]func() any{
				"x": func() any { return 1.0 },
			},
		},
		{
			input:    "x = 1.1",
			expected: false,
			symbols: map[string]func() any{
				"x": func() any { return 1.0 },
			},
		},
		{
			input:    "10 = 1.0 or 1.0 = 10 or 1.0 = 10.0 or 10.0 = 1.0",
			expected: false,
			symbols:  map[string]func() any{},
		},
		{
			input:    "x = true",
			expected: true,
			symbols: map[string]func() any{
				"x": func() any { return true },
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

func TestEvalErrors(t *testing.T) {
	tcs := []struct {
		input    string
		expected error
		symbols  map[string]func() any
	}{
		{
			input: "> y",
			symbols: map[string]func() any{
				"y": func() any { return 5 },
			},
		},
		{
			input:    "x > y",
			expected: ErrSymbolNotFound,
			symbols: map[string]func() any{
				"x": func() any { return 5 },
			},
		},

		{input: `1 = "hello"`, expected: ErrorWrongDataType},
		{input: `1 > "hello"`, expected: ErrorWrongDataType},
		{input: `1 >= "hello"`, expected: ErrorWrongDataType},
		{input: `1 < "hello"`, expected: ErrorWrongDataType},
		{input: `1 <= "hello"`, expected: ErrorWrongDataType},
		{input: `1 != "hello"`, expected: ErrorWrongDataType},

		{input: `1.0 = "hello"`, expected: ErrorWrongDataType},
		{input: `1.0 > "hello"`, expected: ErrorWrongDataType},
		{input: `1.0 >= "hello"`, expected: ErrorWrongDataType},
		{input: `1.0 < "hello"`, expected: ErrorWrongDataType},
		{input: `1.0 <= "hello"`, expected: ErrorWrongDataType},
		{input: `1.0 != "hello"`, expected: ErrorWrongDataType},

		{input: `"hello" = 1`, expected: ErrorWrongDataType},
		{input: `"hello" > 1`, expected: ErrorWrongDataType},
		{input: `"hello" >= 1`, expected: ErrorWrongDataType},
		{input: `"hello" < 1`, expected: ErrorWrongDataType},
		{input: `"hello" <= 1`, expected: ErrorWrongDataType},
		{input: `"hello" != 1`, expected: ErrorWrongDataType},

		{input: `true = 1`, expected: ErrorWrongDataType},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(fmt.Sprintf("%s -> %s", tc.input, tc.expected), func(t *testing.T) {
			_, err := Eval(tc.input, tc.symbols)
			if tc.expected == nil {
				assert.Error(t, err)
			} else {
				assert.ErrorIs(t, err, tc.expected)
			}
		})
	}
}

func TestEvalShortCircuit(t *testing.T) {
	t.Run("short circuit and", func(t *testing.T) {
		input := `x = 0 and y = 0`
		expected := false
		symbols := map[string]func() any{
			"x": func() any { return 1 },
			"y": func() any {
				t.Error("y is called while it shouldn't")
				return 1
			},
		}

		actual, err := Eval(input, symbols)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("short circuit or", func(t *testing.T) {
		input := `x = 0 or y = 0`
		expected := true
		symbols := map[string]func() any{
			"x": func() any { return 0 },
			"y": func() any {
				t.Error("y is called while it shouldn't")
				return 1
			},
		}

		actual, err := Eval(input, symbols)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}
