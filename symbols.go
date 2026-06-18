package boolexpr

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Symbols supplies the value of each symbol referenced by an expression during
// evaluation. Implementations are provided with [SymbolsMap] and
// [SymbolsCached]; custom sources (databases, request context, etc.) can be
// used by implementing this interface.
type Symbols interface {
	// Get returns the resolved value for the named symbol, or an error if the
	// symbol is unknown or its value could not be produced. Errors should wrap
	// [ErrSymbolNotFound] when the symbol is absent.
	Get(string) (any, error)
}

// SymbolsMap is the simplest [Symbols] implementation: a map from symbol name
// to value. A value may be a literal (string, int, float64, bool), a slice for
// use with contains/excludes, or a function that is evaluated lazily on each
// lookup (see [resolveSymbol] for the accepted function signatures).
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

// Used returns the symbols that were actually accessed during evaluation,
// mapped to their resolved values. Because evaluation short-circuits, symbols
// guarded by an already-decided "and"/"or" are absent. Call it after
// evaluating to learn which inputs influenced the result.
func (s *SymbolsCached) Used() map[string]any {
	result := make(map[string]any)

	for k, idx := range s.index {
		if s.entries[idx].used.Load() {
			result[k] = s.entries[idx].val
		}
	}

	return result
}

// NewSymbolsCached builds a [SymbolsCached] from m. The map values follow the
// same rules as [SymbolsMap]: literals, slices, or lazily-evaluated functions.
// Each symbol's value (and any function it wraps) is resolved at most once,
// even across concurrent evaluations.
func NewSymbolsCached(m map[string]any) *SymbolsCached {
	entries := make([]symbolEntry, 0, len(m))
	index := make(map[string]int, len(m))

	for k, v := range m {
		index[k] = len(entries)
		entries = append(entries, symbolEntry{raw: v})
	}

	return &SymbolsCached{index: index, entries: entries}
}

// resolveSymbol turns a raw symbol value into the value used during
// evaluation. Non-function values are returned unchanged. Functions matching
// one of the supported signatures are called and their result used; the
// error-returning variants abort evaluation when they return a non-nil error.
//
// Supported function signatures are, for T in
// {bool, int, string, float64, []string, []int, []float64, []bool, any}:
//
//	func() T
//	func() (T, error)
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
