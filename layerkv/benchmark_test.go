package layerkv

import (
	"context"
	"strconv"
	"testing"

	"github.com/chenyanchen/db"
	"github.com/chenyanchen/db/cachekv"
	"github.com/chenyanchen/db/mocks"
)

// BenchmarkBatch_Set measures the batch Set operation including key extraction.
func BenchmarkBatch_Set(b *testing.B) {
	cache := &mocks.MockBatchKVStore[string, string]{
		DelFunc: func(_ context.Context, _ []string) error {
			return nil
		},
	}
	store := &mocks.MockBatchKVStore[string, string]{
		SetFunc: func(_ context.Context, _ map[string]string) error {
			return nil
		},
	}

	batch, err := NewBatch(cache, store)
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()

	// Create test data of various sizes
	for _, size := range []int{10, 100, 1000} {
		kvs := make(map[string]string, size)
		for i := range size {
			kvs[strconv.Itoa(i)] = "value"
		}

		b.Run("size="+strconv.Itoa(size), func(b *testing.B) {
			b.ReportAllocs()
			for range b.N {
				_ = batch.Set(ctx, kvs)
			}
		})
	}
}

// BenchmarkLayerKV_SetThenGet_WriteInvalidate measures the cost of write-invalidate pattern.
// After Set, the next Get will miss cache and hit store.
func BenchmarkLayerKV_SetThenGet_WriteInvalidate(b *testing.B) {
	cache := cachekv.NewRWMutex[string, string]()
	store := &mocks.MockKVStore[string, string]{
		GetFunc: func(_ context.Context, _ string) (string, error) {
			return "value", nil
		},
		SetFunc: func(_ context.Context, _, _ string) error {
			return nil
		},
		DelFunc: func(_ context.Context, _ string) error {
			return nil
		},
	}

	kv, err := New(cache, store) // Default is WriteInvalidate
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_ = kv.Set(ctx, "key", "value")
		_, _ = kv.Get(ctx, "key") // Will miss cache due to invalidation
	}
}

// BenchmarkLayerKV_SetThenGet_WriteThrough measures the cost of write-through pattern.
// After Set, the next Get will hit cache immediately.
func BenchmarkLayerKV_SetThenGet_WriteThrough(b *testing.B) {
	cache := cachekv.NewRWMutex[string, string]()
	store := &mocks.MockKVStore[string, string]{
		GetFunc: func(_ context.Context, _ string) (string, error) {
			return "value", nil
		},
		SetFunc: func(_ context.Context, _, _ string) error {
			return nil
		},
		DelFunc: func(_ context.Context, _ string) error {
			return nil
		},
	}

	kv, err := New(cache, store, WithWriteThrough())
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_ = kv.Set(ctx, "key", "value")
		_, _ = kv.Get(ctx, "key") // Will hit cache due to write-through
	}
}

// BenchmarkLayerKV_Get_CacheHit measures Get performance when cache hits.
func BenchmarkLayerKV_Get_CacheHit(b *testing.B) {
	cache := cachekv.NewRWMutex[string, string]()
	store := &mocks.MockKVStore[string, string]{
		GetFunc: func(_ context.Context, _ string) (string, error) {
			return "value", nil
		},
	}

	kv, err := New(cache, store)
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()

	// Pre-populate cache
	_ = cache.Set(ctx, "key", "value")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = kv.Get(ctx, "key")
		}
	})
}

// BenchmarkLayerKV_Get_CacheMiss measures Get performance when cache misses.
func BenchmarkLayerKV_Get_CacheMiss(b *testing.B) {
	cache := cachekv.NewRWMutex[string, string]()
	store := &mocks.MockKVStore[string, string]{
		GetFunc: func(_ context.Context, k string) (string, error) {
			return "value", nil
		},
	}

	kv, err := New(cache, store)
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := range b.N {
		// Each iteration uses a new key to force cache miss
		_, _ = kv.Get(ctx, strconv.Itoa(i))
	}
}

// BenchmarkBatch_Get measures batch Get with varying cache hit rates.
func BenchmarkBatch_Get(b *testing.B) {
	const numKeys = 100

	for _, hitRate := range []float64{0.0, 0.5, 1.0} {
		b.Run("hitRate="+strconv.FormatFloat(hitRate, 'f', 1, 64), func(b *testing.B) {
			hitCount := int(float64(numKeys) * hitRate)

			cache := &mocks.MockBatchKVStore[string, string]{
				GetFunc: func(_ context.Context, keys []string) (map[string]string, error) {
					result := make(map[string]string)
					for i, k := range keys {
						if i < hitCount {
							result[k] = "cached"
						}
					}
					return result, nil
				},
				SetFunc: func(_ context.Context, _ map[string]string) error {
					return nil
				},
			}
			store := &mocks.MockBatchKVStore[string, string]{
				GetFunc: func(_ context.Context, keys []string) (map[string]string, error) {
					result := make(map[string]string, len(keys))
					for _, k := range keys {
						result[k] = "stored"
					}
					return result, nil
				},
			}

			batch, err := NewBatch(cache, store)
			if err != nil {
				b.Fatal(err)
			}
			ctx := context.Background()

			keys := make([]string, numKeys)
			for i := range numKeys {
				keys[i] = strconv.Itoa(i)
			}

			b.ResetTimer()
			b.ReportAllocs()
			for range b.N {
				_, _ = batch.Get(ctx, keys)
			}
		})
	}
}

// noopStore is a simple store that does nothing but satisfies db.KV interface.
type noopStore[K comparable, V any] struct {
	val V
}

func (n noopStore[K, V]) Get(_ context.Context, _ K) (V, error) { return n.val, nil }
func (n noopStore[K, V]) Set(_ context.Context, _ K, _ V) error { return nil }
func (n noopStore[K, V]) Del(_ context.Context, _ K) error      { return nil }

// noopBatchStore satisfies db.BatchKV interface.
type noopBatchStore[K comparable, V any] struct{}

func (n noopBatchStore[K, V]) Get(_ context.Context, keys []K) (map[K]V, error) {
	return make(map[K]V), nil
}

func (n noopBatchStore[K, V]) Set(_ context.Context, _ map[K]V) error { return nil }
func (n noopBatchStore[K, V]) Del(_ context.Context, _ []K) error     { return nil }

var (
	_ db.KV[string, string]      = noopStore[string, string]{}
	_ db.BatchKV[string, string] = noopBatchStore[string, string]{}
)
