package boolexpr

import (
	"fmt"
	"sync"

	"github.com/emad-elsaid/memoize"
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

	return v, nil
}

// SymbolsCached implements Symbols interface, wraps map[string]any and keeps track of variables looked up and cache the values returned, gurantees executing any function in the symbols once. suitable for concurrent use
type SymbolsCached struct {
	m         map[string]any
	used      map[string]any
	usedlck   sync.Mutex
	getCached func(string) (any, error)
}

func (s *SymbolsCached) Get(key string) (any, error) {
	return s.getCached(key)
}

// Returns a map of variables used and its values, if the value was a function the map will have the result of the function
func (s *SymbolsCached) Used() map[string]any {
	return s.used
}

// _get is uncached function that gets the value for key and execute the value if it's a function and keep track of the key in `used`
func (s *SymbolsCached) _get(key string) (any, error) {
	v, ok := s.m[key]
	if !ok {
		return v, fmt.Errorf("Symbol: %s, %w", key, ErrSymbolNotFound)
	}

	var resolved any
	var err error

	switch i := v.(type) {
	case func() bool:
		resolved = i()
	case func() (bool, error):
		resolved, err = i()
	case func() int:
		resolved = i()
	case func() (int, error):
		resolved, err = i()
	case func() string:
		resolved = i()
	case func() (string, error):
		resolved, err = i()
	case func() float64:
		resolved = i()
	case func() (float64, error):
		resolved, err = i()
	case func() any:
		resolved = i()
	case func() (any, error):
		resolved, err = i()
	}

	if err != nil {
		return v, fmt.Errorf("Symbol: %s, %w", key, err)
	}

	s.usedlck.Lock()
	s.used[key] = resolved
	s.usedlck.Unlock()

	return v, nil
}

func NewSymbolsCached(m map[string]any) *SymbolsCached {
	s := &SymbolsCached{
		m:    m,
		used: map[string]any{},
	}

	s.getCached = memoize.NewWithErr(s._get)

	return s
}
