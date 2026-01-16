package telemetry

import (
	"context"
	"time"

	kv "github.com/chenyanchen/kv"
)

// RecordFunc is a function that records operation metrics.
type RecordFunc func(operation string, success bool, duration time.Duration)

// KV wraps a kv.KV and records operation metrics.
type KV[K comparable, V any] struct {
	next   kv.KV[K, V]
	record RecordFunc
}

// Wrap wraps a kv.KV with telemetry recording.
func Wrap[K comparable, V any](next kv.KV[K, V], record RecordFunc) *KV[K, V] {
	return &KV[K, V]{next: next, record: record}
}

func (t *KV[K, V]) Get(ctx context.Context, k K) (V, error) {
	start := time.Now()
	v, err := t.next.Get(ctx, k)
	t.record("Get", err == nil, time.Since(start))
	return v, err
}

func (t *KV[K, V]) Set(ctx context.Context, k K, v V) error {
	start := time.Now()
	err := t.next.Set(ctx, k, v)
	t.record("Set", err == nil, time.Since(start))
	return err
}

func (t *KV[K, V]) Del(ctx context.Context, k K) error {
	start := time.Now()
	err := t.next.Del(ctx, k)
	t.record("Del", err == nil, time.Since(start))
	return err
}
