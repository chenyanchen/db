package cachekv

import (
	"context"
	"strconv"
	"sync"
	"testing"

	"github.com/chenyanchen/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShardedKV_Get(t *testing.T) {
	kv := NewSharded[string, string](16)
	ctx := context.Background()

	// Test not found
	_, err := kv.Get(ctx, "missing")
	assert.ErrorIs(t, err, db.ErrNotFound)

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
	assert.ErrorIs(t, err, db.ErrNotFound)

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

func TestNextPowerOf2(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{0, 1},
		{1, 1},
		{2, 2},
		{3, 4},
		{4, 4},
		{5, 8},
		{7, 8},
		{8, 8},
		{9, 16},
		{15, 16},
		{16, 16},
		{17, 32},
		{31, 32},
		{32, 32},
	}

	for _, tt := range tests {
		t.Run(strconv.Itoa(tt.input), func(t *testing.T) {
			assert.Equal(t, tt.expected, nextPowerOf2(tt.input))
		})
	}
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
