package boolexpr

import (
	"fmt"
	"sync"
)

// CachedSymbols wraps a Symbols implementation, caching the results of Get calls
// and tracking which symbols were accessed. Each key is looked up at most once from the
// underlying Symbols implementation. Uses per-entry sync.Once for optimal concurrency.
type CachedSymbols struct {
	underlying Symbols
	mu         sync.Mutex
	entries    map[string]*symbolEntry
}

// NewCachedSymbols creates a new CachedSymbols that wraps the given Symbols implementation.
func NewCachedSymbols(symbols Symbols) *CachedSymbols {
	return &CachedSymbols{
		underlying: symbols,
		entries:    make(map[string]*symbolEntry),
	}
}

// Get returns the value for the given key, caching the result from the underlying Symbols.
// Each key is looked up at most once from the underlying implementation, and the value
// is resolved using resolveSymbol to handle function values.
func (s *CachedSymbols) Get(key string) (any, error) {
	// Get or create the entry for this key
	s.mu.Lock()
	entry, exists := s.entries[key]
	if !exists {
		entry = &symbolEntry{}
		s.entries[key] = entry
	}
	s.mu.Unlock()

	// Use sync.Once to ensure we only fetch and resolve once
	entry.once.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				entry.err = fmt.Errorf("symbol: %s panicked: %v", key, r)
			}
		}()

		// Fetch from underlying Symbols
		rawValue, err := s.underlying.Get(key)
		if err != nil {
			entry.err = err
			return
		}

		// Resolve the symbol (handles functions, etc.)
		resolved, resolveErr := resolveSymbol(rawValue)
		if resolveErr != nil {
			entry.err = resolveErr
			return
		}

		entry.val = resolved
		entry.used.Store(true)
	})

	return entry.val, entry.err
}

// Used returns a map of all symbols that were successfully accessed via Get.
// Symbols that resulted in errors are not included. This matches the behavior of CachedMap.
func (s *CachedSymbols) Used() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make(map[string]any, len(s.entries))
	for key, entry := range s.entries {
		if entry.used.Load() {
			result[key] = entry.val
		}
	}

	return result
}
