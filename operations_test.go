package boolexpr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListSymbols(t *testing.T) {
	exp, err := Parse(`x >= 10 or y < 0 or ( z = "hello" or z = "world" ) and a = b`)
	assert.NoError(t, err)

	actual := ListSymbols(exp)
	expected := []string{"x", "y", "z", "a", "b"}
	assert.ElementsMatch(t, expected, actual)
}
