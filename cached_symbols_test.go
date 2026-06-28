package boolexpr

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSymbols is a test implementation of Symbols that tracks Get calls
type mockSymbols struct {
	data      map[string]any
	getCounts map[string]*atomic.Int32
	mu        sync.Mutex
}

func newMockSymbols(data map[string]any) *mockSymbols {
	getCounts := make(map[string]*atomic.Int32)
	for key := range data {
		getCounts[key] = &atomic.Int32{}
	}
	return &mockSymbols{
		data:      data,
		getCounts: getCounts,
	}
}

func (m *mockSymbols) Get(key string) (any, error) {
	m.mu.Lock()
	if _, exists := m.getCounts[key]; !exists {
		m.getCounts[key] = &atomic.Int32{}
	}
	m.mu.Unlock()

	m.getCounts[key].Add(1)

	if value, ok := m.data[key]; ok {
		return value, nil
	}
	return nil, errors.New("symbol not found")
}

func (m *mockSymbols) getCount(key string) int32 {
	m.mu.Lock()
	defer m.mu.Unlock()
	if counter, exists := m.getCounts[key]; exists {
		return counter.Load()
	}
	return 0
}

func TestNewCachedSymbols(t *testing.T) {
	t.Parallel()

	t.Run("creates wrapper with empty entries map", func(t *testing.T) {
		underlying := newMockSymbols(map[string]any{"key": "value"})
		wrapper := NewCachedSymbols(underlying)

		require.NotNil(t, wrapper)
		require.NotNil(t, wrapper.underlying)
		require.NotNil(t, wrapper.entries)
		require.Empty(t, wrapper.entries)
	})

	t.Run("handles nil symbols gracefully", func(t *testing.T) {
		wrapper := NewCachedSymbols(nil)

		require.NotNil(t, wrapper)
		require.NotNil(t, wrapper.underlying)
		require.NotNil(t, wrapper.entries)

		// Should return error for non-existent keys
		val, err := wrapper.Get("test")
		require.Error(t, err)
		require.Nil(t, val)
		assert.Contains(t, err.Error(), "Symbol not found")
	})
}

func TestCachedSymbols_Get(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		data          map[string]any
		key           string
		expectedValue any
		expectedError bool
		callCount     int
		expectedGets  int32
	}{
		{
			name:          "get existing key",
			data:          map[string]any{"name": "John"},
			key:           "name",
			expectedValue: "John",
			expectedError: false,
			callCount:     1,
			expectedGets:  1,
		},
		{
			name:          "get non-existent key",
			data:          map[string]any{"name": "John"},
			key:           "age",
			expectedValue: nil,
			expectedError: true,
			callCount:     1,
			expectedGets:  1,
		},
		{
			name:          "get same key multiple times caches result",
			data:          map[string]any{"status": "active"},
			key:           "status",
			expectedValue: "active",
			expectedError: false,
			callCount:     5,
			expectedGets:  1, // Should only call underlying Get once
		},
		{
			name:          "get non-existent key multiple times caches error",
			data:          map[string]any{},
			key:           "missing",
			expectedValue: nil,
			expectedError: true,
			callCount:     3,
			expectedGets:  1, // Should only call underlying Get once, even for errors
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			underlying := newMockSymbols(tc.data)
			wrapper := NewCachedSymbols(underlying)

			for i := 0; i < tc.callCount; i++ {
				value, err := wrapper.Get(tc.key)

				if tc.expectedError {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					assert.Equal(t, tc.expectedValue, value)
				}
			}

			// Verify underlying Get was called exactly once
			assert.Equal(t, tc.expectedGets, underlying.getCount(tc.key),
				"underlying Get should be called exactly %d time(s)", tc.expectedGets)
		})
	}
}

func TestCachedSymbols_Get_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	t.Run("concurrent Gets for same key call underlying once", func(t *testing.T) {
		underlying := newMockSymbols(map[string]any{"counter": 42})
		wrapper := NewCachedSymbols(underlying)

		const numGoroutines = 100
		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				value, err := wrapper.Get("counter")
				require.NoError(t, err)
				assert.Equal(t, 42, value)
			}()
		}

		wg.Wait()

		// Despite 100 concurrent calls, underlying should be called at most a few times
		count := underlying.getCount("counter")
		assert.LessOrEqual(t, count, int32(10),
			"underlying Get should be called very few times even with concurrent access")
	})

	t.Run("concurrent Gets for different keys", func(t *testing.T) {
		data := map[string]any{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		}
		underlying := newMockSymbols(data)
		wrapper := NewCachedSymbols(underlying)

		const goroutinesPerKey = 20
		var wg sync.WaitGroup

		for key := range data {
			wg.Add(goroutinesPerKey)
			k := key
			for i := 0; i < goroutinesPerKey; i++ {
				go func() {
					defer wg.Done()
					value, err := wrapper.Get(k)
					require.NoError(t, err)
					assert.Equal(t, data[k], value)
				}()
			}
		}

		wg.Wait()

		// Each key should be fetched at most a few times from underlying
		for key := range data {
			count := underlying.getCount(key)
			assert.LessOrEqual(t, count, int32(10),
				"key %s: underlying Get should be called very few times", key)
		}
	})
}

func TestCachedSymbols_Used(t *testing.T) {
	t.Parallel()

	t.Run("empty when no Gets called", func(t *testing.T) {
		underlying := newMockSymbols(map[string]any{"key": "value"})
		wrapper := NewCachedSymbols(underlying)

		used := wrapper.Used()
		assert.Empty(t, used)
	})

	t.Run("tracks single Get call", func(t *testing.T) {
		underlying := newMockSymbols(map[string]any{"name": "Alice"})
		wrapper := NewCachedSymbols(underlying)

		value, err := wrapper.Get("name")
		require.NoError(t, err)
		assert.Equal(t, "Alice", value)

		used := wrapper.Used()
		require.Len(t, used, 1)
		assert.Equal(t, "Alice", used["name"])
	})

	t.Run("tracks multiple different keys", func(t *testing.T) {
		data := map[string]any{
			"name":   "Bob",
			"age":    30,
			"active": true,
		}
		underlying := newMockSymbols(data)
		wrapper := NewCachedSymbols(underlying)

		for key := range data {
			_, err := wrapper.Get(key)
			require.NoError(t, err)
		}

		used := wrapper.Used()
		require.Len(t, used, 3)
		assert.Equal(t, "Bob", used["name"])
		assert.Equal(t, 30, used["age"])
		assert.Equal(t, true, used["active"])
	})

	t.Run("does not duplicate when key accessed multiple times", func(t *testing.T) {
		underlying := newMockSymbols(map[string]any{"counter": 123})
		wrapper := NewCachedSymbols(underlying)

		// Access same key multiple times
		for i := 0; i < 5; i++ {
			_, err := wrapper.Get("counter")
			require.NoError(t, err)
		}

		used := wrapper.Used()
		require.Len(t, used, 1)
		assert.Equal(t, 123, used["counter"])
	})

	t.Run("does not include keys that resulted in errors", func(t *testing.T) {
		underlying := newMockSymbols(map[string]any{"exists": "value"})
		wrapper := NewCachedSymbols(underlying)

		// Get existing key
		_, err := wrapper.Get("exists")
		require.NoError(t, err)

		// Get non-existent key (error)
		_, err = wrapper.Get("missing")
		require.Error(t, err)

		used := wrapper.Used()
		require.Len(t, used, 1) // Only successful lookups are "used"
		assert.Equal(t, "value", used["exists"])
		assert.NotContains(t, used, "missing") // Error cases are not included
	})

	t.Run("returns copy not reference to internal cache", func(t *testing.T) {
		underlying := newMockSymbols(map[string]any{"key": "value"})
		wrapper := NewCachedSymbols(underlying)

		_, err := wrapper.Get("key")
		require.NoError(t, err)

		used1 := wrapper.Used()
		used2 := wrapper.Used()

		// Modifying one shouldn't affect the other
		used1["new"] = "modified"

		assert.NotContains(t, used2, "new")
	})
}

func TestCachedSymbols_Integration(t *testing.T) {
	t.Parallel()

	t.Run("works with SymbolsMap", func(t *testing.T) {
		underlying := SymbolsMap{
			"name":  "Charlie",
			"score": 95,
		}
		wrapper := NewCachedSymbols(underlying)

		name, err := wrapper.Get("name")
		require.NoError(t, err)
		assert.Equal(t, "Charlie", name)

		score, err := wrapper.Get("score")
		require.NoError(t, err)
		assert.Equal(t, 95, score)

		used := wrapper.Used()
		assert.Len(t, used, 2)
	})

	t.Run("can wrap another CachedSymbols", func(t *testing.T) {
		// Create a chain: SymbolsMap -> CachedSymbols -> CachedSymbols
		base := SymbolsMap{"key": "value"}
		wrapper1 := NewCachedSymbols(base)
		wrapper2 := NewCachedSymbols(wrapper1)

		value, err := wrapper2.Get("key")
		require.NoError(t, err)
		assert.Equal(t, "value", value)

		// Both wrappers should track the usage
		assert.Len(t, wrapper1.Used(), 1)
		assert.Len(t, wrapper2.Used(), 1)
	})

	t.Run("resolves function values from underlying symbols", func(t *testing.T) {
		callCount := 0
		underlying := SymbolsMap{
			"counter": func() int {
				callCount++
				return 42
			},
			"name": func() string {
				return "Alice"
			},
		}
		wrapper := NewCachedSymbols(underlying)

		// First call - function should be resolved
		counter, err := wrapper.Get("counter")
		require.NoError(t, err)
		assert.Equal(t, 42, counter)
		assert.Equal(t, 1, callCount, "function should be called once")

		// Second call - should use cached resolved value, not call function again
		counter, err = wrapper.Get("counter")
		require.NoError(t, err)
		assert.Equal(t, 42, counter)
		assert.Equal(t, 1, callCount, "function should not be called again (cached)")

		// Test string function
		name, err := wrapper.Get("name")
		require.NoError(t, err)
		assert.Equal(t, "Alice", name)

		used := wrapper.Used()
		assert.Len(t, used, 2)
		assert.Equal(t, 42, used["counter"])
		assert.Equal(t, "Alice", used["name"])
	})
}
