package boolexpr

type Symbols interface {
	// returns the value of the key and true if loaded. just like map lookup
	Get(string) (any, bool)
}

type SymbolsMap map[string]any

func (s SymbolsMap) Get(key string) (any, bool) {
	v, ok := s[key]
	return v, ok
}
