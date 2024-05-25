package boolexpr

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	parsec "github.com/prataprc/goparsec"
)

func TestParser(t *testing.T) {
	input := parsec.NewScanner([]byte(`x > 2 and y < 10 or z = "yes" or a = true`))
	node, _ := Parser(input)
	t.Fatalf("Expected no error,val:\n%s got:\nnode: %s",
		"hello",
		spew.Sdump(node),
	)
}
