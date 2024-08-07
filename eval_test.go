package boolexpr

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEval(t *testing.T) {
	tcs := []struct {
		input    string
		expected bool
		symbols  SymbolsMap
	}{
		{
			input:    "x = 1",
			expected: true,
			symbols: map[string]any{
				"x": func() any { return 1 },
			},
		},
		{
			input:    "x = 1",
			expected: true,
			symbols: map[string]any{
				"x": 1,
			},
		},
		{
			input:    "x = 1",
			expected: true,
			symbols: map[string]any{
				"x": 1.0,
			},
		},
		{
			input:    `x = "hello"`,
			expected: true,
			symbols: map[string]any{
				"x": "hello",
			},
		},
		{
			input:    `x = "hello"`,
			expected: false,
			symbols: map[string]any{
				"x": "world",
			},
		},
		{
			input:    `x = true`,
			expected: true,
			symbols: map[string]any{
				"x": true,
			},
		},
		{
			input:    "x = 2",
			expected: false,
			symbols: map[string]any{
				"x": func() any { return 1 },
			},
		},
		{
			input:    `x = "Hello"`,
			expected: true,
			symbols: map[string]any{
				"x": func() any { return "Hello" },
			},
		},
		{
			input:    `x = "Hello"`,
			expected: false,
			symbols: map[string]any{
				"x": func() any { return "World" },
			},
		},
		{
			input:    `x = 2`,
			expected: false,
			symbols: map[string]any{
				"x": func() any { return 2.4 },
			},
		},
		{
			input:    `x = 2`,
			expected: true,
			symbols: map[string]any{
				"x": func() any { return 2.0 },
			},
		},
		{
			input:    `x >= 10`,
			expected: true,
			symbols: map[string]any{
				"x": func() any { return 11 },
			},
		},
		{
			input:    `x != 10`,
			expected: true,
			symbols: map[string]any{
				"x": func() any { return 11 },
			},
		},
		{
			input:    `x <= 10`,
			expected: false,
			symbols: map[string]any{
				"x": func() any { return 11 },
			},
		},
		{
			input:    "x = 1.0",
			expected: true,
			symbols: map[string]any{
				"x": func() any { return 1.0 },
			},
		},
		{
			input:    "x = 1.1",
			expected: false,
			symbols: map[string]any{
				"x": func() any { return 1.0 },
			},
		},
		{
			input:    "10 = 1.0 or 1.0 = 10 or 1.0 = 10.0 or 10.0 = 1.0",
			expected: false,
			symbols:  map[string]any{},
		},
		{
			input:    "x = true",
			expected: true,
			symbols: map[string]any{
				"x": func() any { return true },
			},
		},
		{
			input:    "x = true",
			expected: true,
			symbols: map[string]any{
				"x": func() (any, error) { return true, nil },
			},
		},
		{
			input:    `x >= 10 and y < 0`,
			expected: true,
			symbols: map[string]any{
				"x": func() any { return 11 },
				"y": func() any { return -1 },
			},
		},
		{
			input:    `x >= 10 and y < 0`,
			expected: false,
			symbols: map[string]any{
				"x": func() any { return 11 },
				"y": func() any { return 0 },
			},
		},
		{
			input:    `x >= 10 or y < 0`,
			expected: true,
			symbols: map[string]any{
				"x": func() any { return 11 },
				"y": func() any { return 0 },
			},
		},
		{
			input:    `x >= 10 or y < 0 or ( z = "hello" or z = "world" )`,
			expected: true,
			symbols: map[string]any{
				"x": func() any { return 0 },
				"y": func() any { return 0 },
				"z": func() any { return "hello" },
			},
		},
		{
			input:    `x >= 10 or y < 0 or ( z = "hello" or z = "world" )`,
			expected: false,
			symbols: map[string]any{
				"x": func() any { return 0 },
				"y": func() any { return 0 },
				"z": func() any { return "NO" },
			},
		},
		{
			input:    `x > y and y > z`,
			expected: true,
			symbols: map[string]any{
				"x": func() any { return 10 },
				"y": func() any { return 5 },
				"z": func() any { return 2 },
			},
		},
		{
			input:    `x > y and y > z`,
			expected: false,
			symbols: map[string]any{
				"x": func() any { return 10 },
				"y": func() any { return 5 },
				"z": func() any { return 6 },
			},
		},
		{input: `1 > 0.9`, expected: true},
		{input: `1.1 > 1`, expected: true},
		{input: `1.1 > 1.0`, expected: true},
		{input: `"AB" > "AA"`, expected: true},

		{input: `1 >= 0.9`, expected: true},
		{input: `1.1 >= 1`, expected: true},
		{input: `1.1 >= 1.0`, expected: true},
		{input: `"AB" >= "AA"`, expected: true},

		{input: `0.9 < 1`, expected: true},
		{input: `1 < 1.1`, expected: true},
		{input: `1 < 1.1`, expected: true},
		{input: `1.0 < 1.1`, expected: true},
		{input: `"AA" < "AB"`, expected: true},

		{input: `0.9 <= 1`, expected: true},
		{input: `1 <= 1.1`, expected: true},
		{input: `1 <= 1.1`, expected: true},
		{input: `1.0 <= 1.1`, expected: true},
		{input: `"AA" <= "AB"`, expected: true},

		{input: `0.9 != 1`, expected: true},
		{input: `1 != 1.1`, expected: true},
		{input: `1 != 1.1`, expected: true},
		{input: `1.0 != 1.1`, expected: true},
		{input: `"AA" != "AB"`, expected: true},
		{input: `true != false`, expected: true},
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
		symbols  SymbolsMap
	}{
		{
			input: "> y",
			symbols: map[string]any{
				"y": func() any { return 5 },
			},
		},
		{
			input:    "x != true or y < 0",
			expected: ErrSymbolNotFound,
			symbols: map[string]any{
				"x": func() any { return true },
			},
		},
		{
			input:    "x = x and ( x > y )",
			expected: ErrSymbolNotFound,
			symbols: map[string]any{
				"x": func() any { return 5 },
			},
		},
		{
			input:    "x = x and ( x > y )",
			expected: io.ErrShortBuffer,
			symbols: map[string]any{
				"x": func() (any, error) { return 5, io.ErrShortBuffer },
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
		{input: `true != 1`, expected: ErrorWrongDataType},
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
		symbols := map[string]any{
			"x": func() any { return 1 },
			"y": func() any {
				t.Error("y is called while it shouldn't")
				return 1
			},
		}

		actual, err := Eval(input, SymbolsMap(symbols))
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("short circuit or", func(t *testing.T) {
		input := `x = 0 or y = 0`
		expected := true
		symbols := map[string]any{
			"x": func() any { return 0 },
			"y": func() any {
				t.Error("y is called while it shouldn't")
				return 1
			},
		}

		actual, err := Eval(input, SymbolsMap(symbols))
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}
