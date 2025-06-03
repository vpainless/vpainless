package collect

import (
	"iter"
	"sync"
	"sync/atomic"
)

// Map is a generic map safe to be used concurrently.
type Map[K comparable, V any] struct {
	m      sync.Map
	length atomic.Int32
}

// Stores the provided value for the specified key.
func (m *Map[K, V]) Store(key K, value V) {
	if _, ok := m.m.Load(key); !ok {
		m.length.Add(1)
	}

	m.m.Store(key, value)
}

// Loads the values stored for the specified key,
// and returns true if it exists.
func (m *Map[K, V]) Load(key K) (V, bool) {
	val, ok := m.m.Load(key)
	if !ok {
		var zero V
		return zero, false
	}
	return val.(V), true
}

// Deletes the value for the specified key.
func (m *Map[K, V]) Delete(key K) {
	if _, ok := m.m.Load(key); ok {
		m.length.Add(-1)
	}

	m.m.Delete(key)
}

// Len returns the number of items stored in the Map
func (m *Map[K, V]) Len() int {
	return int(m.length.Load())
}

func (m *Map[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(k K, v V) bool) {
		m.m.Range(func(key, value any) bool {
			return yield(key.(K), value.(V))
		})
	}
}
