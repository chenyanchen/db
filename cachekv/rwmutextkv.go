package cachekv

import (
	"context"
	"sync"

	kv "github.com/chenyanchen/kv"
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
	v, ok := s.m[k]
	s.mu.RUnlock()

	if !ok {
		return v, kv.ErrNotFound
	}
	return v, nil
}

func (s *rwMutexKV[K, V]) Set(ctx context.Context, k K, v V) error {
	s.mu.Lock()
	s.m[k] = v
	s.mu.Unlock()
	return nil
}

func (s *rwMutexKV[K, V]) Del(ctx context.Context, k K) error {
	s.mu.Lock()
	delete(s.m, k)
	s.mu.Unlock()
	return nil
}
