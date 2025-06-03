package collect

import (
	"sync"
	"sync/atomic"
)

// Map is a generic set safe to be used concurrently.
type Set[K comparable] struct {
	m      sync.Map
	length atomic.Int32
}

// Add inserts an element into the set.
func (s *Set[K]) Add(key K) {
	if _, ok := s.m.Load(key); ok {
		return
	}
	s.length.Add(1)
	s.m.Store(key, struct{}{})
}

// Contains checks if the set contains the specifier element.
func (s *Set[K]) Contains(key K) bool {
	_, ok := s.m.Load(key)
	return ok
}

// Deletes the specified element in the set.
func (s *Set[K]) Delete(key K) {
	if _, ok := s.m.Load(key); ok {
		s.length.Add(-1)
	}

	s.m.Delete(key)
}

// Len returns the number of elements in the set.
func (s *Set[K]) Len() int {
	return int(s.length.Load())
}
