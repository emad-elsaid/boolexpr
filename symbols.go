package boolexpr

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Symbols is an interface that returns the value for symbols for evaluator
type Symbols interface {
	// returns the value of the key and error if not found or other reasons
	Get(string) (any, error)
}

// SymbolsMap is a simple wrapper around map[string]any
type SymbolsMap map[string]any

func (s SymbolsMap) Get(key string) (any, error) {
	v, ok := s[key]
	if !ok {
		return v, fmt.Errorf("Symbol: %s, %w", key, ErrSymbolNotFound)
	}

	resolved, err := resolveSymbol(v)
	if err != nil {
		return nil, fmt.Errorf("Symbol: %s, %w", key, err)
	}

	return resolved, nil
}

type symbolEntry struct {
	once sync.Once
	raw  any
	val  any
	err  error
	used atomic.Bool
}

// SymbolsCached implements Symbols interface, wraps map[string]any and keeps
// track of variables looked up and caches the values returned, guarantees
// executing any function in the symbols once. Suitable for concurrent use.
//
// Entries are stored in a contiguous slice; the index map holds name→slice-index
// to avoid N individual heap pointer allocations that map[string]*symbolEntry would require.
type SymbolsCached struct {
	index   map[string]int
	entries []symbolEntry
}

func (s *SymbolsCached) Get(key string) (any, error) {
	idx, ok := s.index[key]
	if !ok {
		return nil, fmt.Errorf("symbol: %s, %w", key, ErrSymbolNotFound)
	}

	e := &s.entries[idx]
	e.once.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				e.err = fmt.Errorf("symbol: %s panicked: %v", key, r)
			}
		}()

		resolved, err := resolveSymbol(e.raw)
		if err != nil {
			e.err = fmt.Errorf("symbol: %s, %w", key, err)
			return
		}

		e.val = resolved
		e.used.Store(true)
	})

	return e.val, e.err
}

// Used returns a map of variables that were accessed and their resolved values.
func (s *SymbolsCached) Used() map[string]any {
	result := make(map[string]any)

	for k, idx := range s.index {
		if s.entries[idx].used.Load() {
			result[k] = s.entries[idx].val
		}
	}

	return result
}

func NewSymbolsCached(m map[string]any) *SymbolsCached {
	entries := make([]symbolEntry, 0, len(m))
	index := make(map[string]int, len(m))

	for k, v := range m {
		index[k] = len(entries)
		entries = append(entries, symbolEntry{raw: v})
	}

	return &SymbolsCached{index: index, entries: entries}
}

func resolveSymbol(v any) (any, error) {
	switch i := v.(type) {
	case func() bool:
		return i(), nil
	case func() (bool, error):
		return i()
	case func() int:
		return i(), nil
	case func() (int, error):
		return i()
	case func() string:
		return i(), nil
	case func() (string, error):
		return i()
	case func() float64:
		return i(), nil
	case func() (float64, error):
		return i()
	case func() []string:
		return i(), nil
	case func() ([]string, error):
		return i()
	case func() []int:
		return i(), nil
	case func() ([]int, error):
		return i()
	case func() []float64:
		return i(), nil
	case func() ([]float64, error):
		return i()
	case func() []bool:
		return i(), nil
	case func() ([]bool, error):
		return i()
	case func() any:
		return i(), nil
	case func() (any, error):
		return i()
	default:
		return i, nil
	}
}
