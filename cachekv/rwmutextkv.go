package cachekv

import (
	"context"
	"sync"

	"github.com/chenyanchen/db"
)

type rwMutexKV[K comparable, V any] struct {
	mu sync.RWMutex
	kv map[K]V
}

func NewRWMutex[K comparable, V any]() *rwMutexKV[K, V] {
	return &rwMutexKV[K, V]{kv: make(map[K]V)}
}

func (s *rwMutexKV[K, V]) Get(ctx context.Context, k K) (V, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.kv[k]
	if !ok {
		return v, db.ErrNotFound
	}
	return v, nil
}

func (s *rwMutexKV[K, V]) Set(ctx context.Context, k K, v V) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.kv[k] = v
	return nil
}

func (s *rwMutexKV[K, V]) Del(ctx context.Context, k K) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.kv, k)
	return nil
}
