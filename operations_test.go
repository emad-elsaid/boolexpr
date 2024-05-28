package boolexpr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListSymbols(t *testing.T) {
	exp, err := Parse(`
a >= b or
c < d or
( e = "hello" or f = "world" ) and
g = h or
a = b`,
	)
	assert.NoError(t, err)

	actual := ListSymbols(exp)
	expected := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	assert.ElementsMatch(t, expected, actual)
}
