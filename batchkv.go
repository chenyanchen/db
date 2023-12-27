package db

import "context"

// BatchKV is a batch key-value storage.
type BatchKV[K comparable, V any] interface {
	Get(ctx context.Context, keys []K) (map[K]V, error)
	Set(ctx context.Context, kvs map[K]V) error
	Del(ctx context.Context, keys []K) error
}
