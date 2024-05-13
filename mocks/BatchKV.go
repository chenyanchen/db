package mocks

import "context"

type MockBatchKVStore[K comparable, V any] struct {
	GetFunc func(context.Context, []K) (map[K]V, error)
	SetFunc func(context.Context, map[K]V) error
	DelFunc func(context.Context, []K) error
}

func (s MockBatchKVStore[K, V]) Get(ctx context.Context, keys []K) (map[K]V, error) {
	return s.GetFunc(ctx, keys)
}

func (s MockBatchKVStore[K, V]) Set(ctx context.Context, kvs map[K]V) error {
	return s.SetFunc(ctx, kvs)
}

func (s MockBatchKVStore[K, V]) Del(ctx context.Context, keys []K) error {
	return s.DelFunc(ctx, keys)
}
