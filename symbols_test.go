package boolexpr

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSymbolsMap(t *testing.T) {
	t.Run("implements Symbols", func(t *testing.T) {
		var _ Symbols = SymbolsMap{}
	})
}

func TestSymbolsCached(t *testing.T) {
	t.Run("implements Symbols", func(t *testing.T) {
		var _ Symbols = NewSymbolsCached(nil)
	})

	t.Run("Works for direct values", func(t *testing.T) {
		s := NewSymbolsCached(map[string]any{
			"x": 1,
			"y": 2,
		})

		res, err := Eval("x = 1 and x != 0 and y = 2 and y != 0", s)
		assert.NoError(t, err)
		assert.True(t, res)

		expected := map[string]any{"x": 1, "y": 2}
		assert.Equal(t, expected, s.Used())
	})

	t.Run("eval with variables with values mixed (funcs, literals)", func(t *testing.T) {
		xcalled := 0
		ycalled := 0
		s := NewSymbolsCached(map[string]any{
			"x": func() int {
				xcalled++
				return 1
			},
			"y": func() int {
				ycalled++
				return 2
			},
		})

		res, err := Eval("x = 1 and x != 0 and y = 2 and y != 0", s)
		assert.NoError(t, err)
		assert.True(t, res)

		expected := map[string]any{"x": 1, "y": 2}
		assert.Equal(t, expected, s.Used())

		assert.Equal(t, 1, xcalled)
		assert.Equal(t, 1, ycalled)
	})

	t.Run("keep track of used variables", func(t *testing.T) {
		s := NewSymbolsCached(map[string]any{
			"x": func() int { return 1 },
			"y": func() int { return 2 },
		})

		res, err := Eval("x = 0 and y = 0", s)
		assert.NoError(t, err)
		assert.False(t, res)

		expected := map[string]any{"x": 1}
		assert.Equal(t, expected, s.Used())
	})

	t.Run("i a function returned error, still records symbols", func(t *testing.T) {
		s := NewSymbolsCached(map[string]any{
			"x": func() int { return 1 },
			"y": func() int { return 2 },
			"z": func() (int, error) { return 3, fmt.Errorf("Z errored") },
		})

		res, err := Eval("x = 1 and y = 2 and z = 3", s)
		assert.Error(t, err)
		assert.False(t, res)

		expected := map[string]any{"x": 1, "y": 2}
		assert.Equal(t, expected, s.Used())
	})
}
