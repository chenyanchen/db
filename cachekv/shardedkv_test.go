package cachekv

import (
	"context"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	kvpkg "github.com/chenyanchen/kv"
)

func TestShardedKV_Get(t *testing.T) {
	kv := NewSharded[string, string](16)
	ctx := context.Background()

	// Test not found
	_, err := kv.Get(ctx, "missing")
	require.ErrorIs(t, err, kvpkg.ErrNotFound)

	// Test found
	require.NoError(t, kv.Set(ctx, "key1", "value1"))
	v, err := kv.Get(ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", v)
}

func TestShardedKV_Set(t *testing.T) {
	kv := NewSharded[string, string](16)
	ctx := context.Background()

	// Set and retrieve
	require.NoError(t, kv.Set(ctx, "key1", "value1"))
	v, err := kv.Get(ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", v)

	// Overwrite
	require.NoError(t, kv.Set(ctx, "key1", "value2"))
	v, err = kv.Get(ctx, "key1")
	require.NoError(t, err)
	assert.Equal(t, "value2", v)
}

func TestShardedKV_Del(t *testing.T) {
	kv := NewSharded[string, string](16)
	ctx := context.Background()

	// Set then delete
	require.NoError(t, kv.Set(ctx, "key1", "value1"))
	require.NoError(t, kv.Del(ctx, "key1"))

	// Should be not found
	_, err := kv.Get(ctx, "key1")
	require.ErrorIs(t, err, kvpkg.ErrNotFound)

	// Delete non-existent is fine
	require.NoError(t, kv.Del(ctx, "missing"))
}

func TestShardedKV_Len(t *testing.T) {
	kv := NewSharded[string, string](16)
	ctx := context.Background()

	assert.Equal(t, 0, kv.Len())

	for i := range 100 {
		require.NoError(t, kv.Set(ctx, strconv.Itoa(i), "value"))
	}
	assert.Equal(t, 100, kv.Len())

	for i := range 50 {
		require.NoError(t, kv.Del(ctx, strconv.Itoa(i)))
	}
	assert.Equal(t, 50, kv.Len())
}

func TestShardedKV_IntKey(t *testing.T) {
	kv := NewSharded[int, string](16)
	ctx := context.Background()

	require.NoError(t, kv.Set(ctx, 42, "answer"))
	v, err := kv.Get(ctx, 42)
	require.NoError(t, err)
	assert.Equal(t, "answer", v)
}

func TestShardedKV_StructKey(t *testing.T) {
	type Point struct{ X, Y int }

	kv := NewSharded[Point, string](16)
	ctx := context.Background()

	// Set with one variable
	p1 := Point{X: 1, Y: 2}
	require.NoError(t, kv.Set(ctx, p1, "origin"))

	// Get with a different variable (same value)
	p2 := Point{X: 1, Y: 2}
	v, err := kv.Get(ctx, p2)
	require.NoError(t, err)
	assert.Equal(t, "origin", v)

	// Overwrite works
	require.NoError(t, kv.Set(ctx, p2, "updated"))
	v, err = kv.Get(ctx, p1)
	require.NoError(t, err)
	assert.Equal(t, "updated", v)
}

func TestShardedKV_Concurrent(t *testing.T) {
	kv := NewSharded[string, int](32)
	ctx := context.Background()

	const numGoroutines = 100
	const numOps = 1000

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for g := range numGoroutines {
		go func(id int) {
			defer wg.Done()
			for i := range numOps {
				key := strconv.Itoa((id*numOps + i) % 1000)
				switch i % 3 {
				case 0:
					_ = kv.Set(ctx, key, i)
				case 1:
					_, _ = kv.Get(ctx, key)
				case 2:
					_ = kv.Del(ctx, key)
				}
			}
		}(g)
	}

	wg.Wait()
}

// Benchmarks comparing sharded vs non-sharded

func BenchmarkShardedKV_Get(b *testing.B) {
	for _, shards := range []int{16, 32, 64} {
		b.Run("shards="+strconv.Itoa(shards), func(b *testing.B) {
			kv := NewSharded[string, string](shards)
			ctx := context.Background()

			// Pre-populate
			for i := range benchKeyCount {
				_ = kv.Set(ctx, strconv.Itoa(i), "value")
			}

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					_, _ = kv.Get(ctx, strconv.Itoa(i%benchKeyCount))
					i++
				}
			})
		})
	}
}

func BenchmarkShardedKV_Set(b *testing.B) {
	for _, shards := range []int{16, 32, 64} {
		b.Run("shards="+strconv.Itoa(shards), func(b *testing.B) {
			kv := NewSharded[string, string](shards)
			ctx := context.Background()

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					_ = kv.Set(ctx, strconv.Itoa(i%benchKeyCount), "value")
					i++
				}
			})
		})
	}
}

func BenchmarkShardedKV_Mixed(b *testing.B) {
	for _, shards := range []int{16, 32, 64} {
		b.Run("shards="+strconv.Itoa(shards), func(b *testing.B) {
			kv := NewSharded[string, string](shards)
			ctx := context.Background()

			// Pre-populate
			for i := range benchKeyCount {
				_ = kv.Set(ctx, strconv.Itoa(i), "value")
			}

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					key := strconv.Itoa(i % benchKeyCount)
					if i%5 == 0 {
						_ = kv.Set(ctx, key, "value")
					} else {
						_, _ = kv.Get(ctx, key)
					}
					i++
				}
			})
		})
	}
}
