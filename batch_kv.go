package db

import "context"

// BatchKV represent a key-val storage to store values.
type BatchKV[K comparable, V any] interface {
	Get(context.Context, []K) (map[K]V, error)
	Set(context.Context, map[K]V) error
	Del(context.Context, []K) error
}
