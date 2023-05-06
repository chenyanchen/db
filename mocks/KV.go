package mocks

import "context"

type MockKVStore[K comparable, V any] struct {
	GetFunc func(context.Context, K) (V, error)
	SetFunc func(context.Context, K, V) error
	DelFunc func(context.Context, K) error
}

func (s MockKVStore[K, V]) Get(ctx context.Context, k K) (V, error) { return s.GetFunc(ctx, k) }
func (s MockKVStore[K, V]) Set(ctx context.Context, k K, v V) error { return s.SetFunc(ctx, k, v) }
func (s MockKVStore[K, V]) Del(ctx context.Context, k K) error      { return s.DelFunc(ctx, k) }
