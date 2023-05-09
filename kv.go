package db

import (
	"context"
	"errors"
)

// KV represent a key-val storage to store values.
type KV[K comparable, V any] interface {
	Get(context.Context, K) (V, error)
	Set(context.Context, K, V) error
	Del(context.Context, K) error
}

var NotFound = errors.New("not found")
