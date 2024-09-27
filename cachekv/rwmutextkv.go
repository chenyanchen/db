package cachekv

import (
	"context"
	"sync"

	"github.com/chenyanchen/db"
)

type rwMutexKV[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

func NewRWMutex[K comparable, V any]() *rwMutexKV[K, V] {
	return &rwMutexKV[K, V]{m: make(map[K]V)}
}

func (s *rwMutexKV[K, V]) Get(ctx context.Context, k K) (V, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.m[k]
	if !ok {
		return v, db.ErrNotFound
	}
	return v, nil
}

func (s *rwMutexKV[K, V]) Set(ctx context.Context, k K, v V) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.m[k] = v
	return nil
}

func (s *rwMutexKV[K, V]) Del(ctx context.Context, k K) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.m, k)
	return nil
}
