package boolexpr

import "testing"

func TestSymbolsMap(t *testing.T) {
	t.Run("implements Symbols", func(t *testing.T) {
		var _ Symbols = SymbolsMap{}
	})
}
