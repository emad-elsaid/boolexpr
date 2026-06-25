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
			input:    "x == 1",
			expected: true,
			symbols: map[string]any{
				"x": func() any { return 1 },
			},
		},
		{
			input:    "x == 1",
			expected: true,
			symbols: map[string]any{
				"x": 1,
			},
		},
		{
			input:    "x == 1",
			expected: true,
			symbols: map[string]any{
				"x": 1.0,
			},
		},
		{
			input:    `x == "hello"`,
			expected: true,
			symbols: map[string]any{
				"x": "hello",
			},
		},
		{
			input:    `x == "hello"`,
			expected: false,
			symbols: map[string]any{
				"x": "world",
			},
		},
		{
			input:    `x == true`,
			expected: true,
			symbols: map[string]any{
				"x": true,
			},
		},
		{
			input:    "x == 2",
			expected: false,
			symbols: map[string]any{
				"x": func() any { return 1 },
			},
		},
		{
			input:    `x == "Hello"`,
			expected: true,
			symbols: map[string]any{
				"x": func() any { return "Hello" },
			},
		},
		{
			input:    `x == "Hello"`,
			expected: false,
			symbols: map[string]any{
				"x": func() any { return "World" },
			},
		},
		{
			input:    `x == 2`,
			expected: false,
			symbols: map[string]any{
				"x": func() any { return 2.4 },
			},
		},
		{
			input:    `x == 2`,
			expected: true,
			symbols: map[string]any{
				"x": func() any { return 2.0 },
			},
		},
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
		{input: `1 = 1 && 2 = 2`, expected: true},
		{input: `1 = 1 && 2 = 3`, expected: false},
		{input: `1 = 2 || 2 = 2`, expected: true},
		{input: `1 = 2 || 2 = 3`, expected: false},
		{input: `1 = 1 && 2 = 2 || 3 = 4`, expected: true},
		{input: `1 = 2 && 2 = 2 || ( 3 = 3 )`, expected: true},

		// Operator precedence: "and" binds tighter than "or", matching Go.
		// Each case below flips result if evaluated flat left-to-right
		// (the previous behavior) instead of `a or (b and c)`, so they pin
		// the precedence down. The parenthesized twin forces the flat reading
		// and must give the opposite result.
		{input: `true or false and false`, expected: true},   // true or (false and false)
		{input: `( true or false ) and false`, expected: false}, // flat reading -> false
		{input: `true or true and false`, expected: true},    // true or (true and false)
		{input: `( true or true ) and false`, expected: false},
		{input: `true and true or false and false`, expected: true}, // (T and T) or (F and F)
		{input: `true and ( true or false ) and false`, expected: false},

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

		// bare bool value
		{input: `true`, expected: true},
		{input: `false`, expected: false},
		{
			input:    `x`,
			expected: true,
			symbols:  SymbolsMap{"x": true},
		},
		{
			input:    `x`,
			expected: false,
			symbols:  SymbolsMap{"x": false},
		},
		{
			input:    `x and y`,
			expected: true,
			symbols:  SymbolsMap{"x": true, "y": true},
		},
		{
			input:    `x and y = 1`,
			expected: true,
			symbols:  SymbolsMap{"x": true, "y": 1},
		},

		// contains: string contains substring
		{input: `"hello world" contains "world"`, expected: true},
		{input: `"hello world" contains "xyz"`, expected: false},
		{input: `"hello world" contains ""`, expected: true},

		// contains: []string variable contains a string literal
		{
			input:    `tags contains "go"`,
			expected: true,
			symbols:  SymbolsMap{"tags": []string{"go", "rust", "c"}},
		},
		{
			input:    `tags contains "java"`,
			expected: false,
			symbols:  SymbolsMap{"tags": []string{"go", "rust", "c"}},
		},

		// contains: []int variable
		{
			input:    `ids contains 42`,
			expected: true,
			symbols:  SymbolsMap{"ids": []int{1, 42, 99}},
		},
		{
			input:    `ids contains 0`,
			expected: false,
			symbols:  SymbolsMap{"ids": []int{1, 42, 99}},
		},
		// int/float64 cross-compatibility
		{
			input:    `ids contains 42.0`,
			expected: true,
			symbols:  SymbolsMap{"ids": []int{1, 42, 99}},
		},
		{
			input:    `ids contains 0.0`,
			expected: false,
			symbols:  SymbolsMap{"ids": []int{1, 42, 99}},
		},
		// contains: []float64 variable
		{
			input:    `scores contains 3.0`,
			expected: true,
			symbols:  SymbolsMap{"scores": []float64{1.5, 3.0, 9.9}},
		},
		{
			input:    `scores contains 2.0`,
			expected: false,
			symbols:  SymbolsMap{"scores": []float64{1.5, 3.0, 9.9}},
		},
		{
			input:    `scores contains 3`,
			expected: true,
			symbols:  SymbolsMap{"scores": []float64{1.5, 3.0, 9.9}},
		},
		{
			input:    `scores contains 2`,
			expected: false,
			symbols:  SymbolsMap{"scores": []float64{1.5, 3.0, 9.9}},
		},
		// contains: []bool variable
		{
			input:    `flags contains true`,
			expected: true,
			symbols:  SymbolsMap{"flags": []bool{false, true}},
		},
		{
			input:    `flags contains true`,
			expected: false,
			symbols:  SymbolsMap{"flags": []bool{false, false}},
		},

		// excludes: negation of contains
		{input: `"hello world" excludes "xyz"`, expected: true},
		{input: `"hello world" excludes "world"`, expected: false},
		{
			input:    `tags excludes "java"`,
			expected: true,
			symbols:  SymbolsMap{"tags": []string{"go", "rust"}},
		},
		{
			input:    `tags excludes "go"`,
			expected: false,
			symbols:  SymbolsMap{"tags": []string{"go", "rust"}},
		},
		{
			input:    `ids excludes 0`,
			expected: true,
			symbols:  SymbolsMap{"ids": []int{1, 2, 3}},
		},
		{
			input:    `ids excludes 1`,
			expected: false,
			symbols:  SymbolsMap{"ids": []int{1, 2, 3}},
		},
		{
			input:    `scores excludes 9.9`,
			expected: false,
			symbols:  SymbolsMap{"scores": []float64{1.5, 9.9}},
		},
		{
			input:    `scores excludes 2.0`,
			expected: true,
			symbols:  SymbolsMap{"scores": []float64{1.5, 9.9}},
		},
		{
			input:    `flags excludes false`,
			expected: true,
			symbols:  SymbolsMap{"flags": []bool{true, true}},
		},
		{
			input:    `flags excludes false`,
			expected: false,
			symbols:  SymbolsMap{"flags": []bool{true, false}},
		},

		// starts_with
		{input: `"hello world" starts_with "hello"`, expected: true},
		{input: `"hello world" starts_with "world"`, expected: false},
		{input: `"hello world" starts_with ""`, expected: true},
		{
			input:    `name starts_with "Jo"`,
			expected: true,
			symbols:  SymbolsMap{"name": "John"},
		},

		// ends_with
		{input: `"hello world" ends_with "world"`, expected: true},
		{input: `"hello world" ends_with "hello"`, expected: false},
		{input: `"hello world" ends_with ""`, expected: true},
		{
			input:    `name ends_with "hn"`,
			expected: true,
			symbols:  SymbolsMap{"name": "John"},
		},

		// match
		{input: `"pattern123" match "pattern.*"`, expected: true},
		{input: `"abc" match "^abc$"`, expected: true},
		{input: `"abcd" match "^abc$"`, expected: false},
		{
			input:    `x match "pattern.*"`,
			expected: true,
			symbols:  SymbolsMap{"x": "pattern123"},
		},
		{
			input:    `email match pattern`,
			expected: true,
			symbols:  SymbolsMap{"email": "joanna@example.com", "pattern": `.+@example\.com$`},
		},
		{
			input:    `x match p`,
			expected: true,
			symbols: SymbolsMap{
				"x": func() string { return "foo42" },
				"p": func() string { return "[0-9]+" },
			},
		},

		// Exact integer comparisons beyond 2^53, where float64 coercion would
		// collapse distinct integers. 9007199254740992 is 2^53.
		{
			input:    `x = 9007199254740992`,
			expected: false,
			symbols:  SymbolsMap{"x": 9007199254740993},
		},
		{
			input:    `x != 9007199254740992`,
			expected: true,
			symbols:  SymbolsMap{"x": 9007199254740993},
		},
		{
			input:    `x > 9007199254740992`,
			expected: true,
			symbols:  SymbolsMap{"x": 9007199254740993},
		},
		{
			input:    `x >= 9007199254740993`,
			expected: true,
			symbols:  SymbolsMap{"x": 9007199254740993},
		},
		{
			input:    `x < 9007199254740994`,
			expected: true,
			symbols:  SymbolsMap{"x": 9007199254740993},
		},
		// Both operands are integer symbols just past 2^53.
		{
			input:    `x = y`,
			expected: false,
			symbols:  SymbolsMap{"x": 9007199254740993, "y": 9007199254740992},
		},
		// []int membership must use exact integer comparison.
		{
			input:    `ids contains 9007199254740992`,
			expected: false,
			symbols:  SymbolsMap{"ids": []int{9007199254740993}},
		},
		{
			input:    `ids excludes 9007199254740992`,
			expected: true,
			symbols:  SymbolsMap{"ids": []int{9007199254740993}},
		},
		{
			input:    `ids contains 9007199254740993`,
			expected: true,
			symbols:  SymbolsMap{"ids": []int{9007199254740993}},
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

		// bare non-bool value
		{
			input:    `x`,
			expected: ErrorWrongDataType,
			symbols:  SymbolsMap{"x": 42},
		},

		// excludes: type mismatch propagates from containsEval. The result must
		// stay false on error — excludes negates containsEval, so a careless
		// error path would flip false into true.
		{input: `1 excludes "x"`, expected: ErrorWrongDataType},
		{
			input:    `tags excludes 1`,
			expected: ErrorWrongDataType,
			symbols:  SymbolsMap{"tags": []string{"a"}},
		},

		// contains: type mismatches
		{input: `1 contains "x"`, expected: ErrorWrongDataType},
		{input: `"hello" contains 1`, expected: ErrorWrongDataType},
		{
			input:    `scores contains "x"`,
			expected: ErrorWrongDataType,
			symbols:  SymbolsMap{"scores": []float64{1.5, 3.0}},
		},
		{
			input:    `tags contains 1`,
			expected: ErrorWrongDataType,
			symbols:  SymbolsMap{"tags": []string{"a", "b"}},
		},
		{
			input:    `ids contains "x"`,
			expected: ErrorWrongDataType,
			symbols:  SymbolsMap{"ids": []int{1, 2}},
		},
		{
			input:    `flags contains 1`,
			expected: ErrorWrongDataType,
			symbols:  SymbolsMap{"flags": []bool{true, false}},
		},

		// starts_with / ends_with type mismatches
		{input: `1 starts_with "x"`, expected: ErrorWrongDataType},
		{input: `"hello" starts_with 1`, expected: ErrorWrongDataType},
		{input: `1 ends_with "x"`, expected: ErrorWrongDataType},
		{input: `"hello" ends_with 1`, expected: ErrorWrongDataType},

		// match type mismatches and invalid pattern
		{input: `1 match "x"`, expected: ErrorWrongDataType},
		{input: `"hello" match 1`, expected: ErrorWrongDataType},
		{input: `"hello" match "("`, expected: ErrorWrongDataType},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(fmt.Sprintf("%s -> %s", tc.input, tc.expected), func(t *testing.T) {
			output, err := Eval(tc.input, tc.symbols)
			if tc.expected == nil {
				assert.Error(t, err)
			} else {
				assert.ErrorIs(t, err, tc.expected)
			}
			assert.False(t, output, "result must be false when evaluation errors")
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
