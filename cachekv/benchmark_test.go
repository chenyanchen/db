package cachekv

import (
	"context"
	"strconv"
	"testing"
)

const benchKeyCount = 10000

// BenchmarkRWMutexKV_Get benchmarks Get under parallel access.
func BenchmarkRWMutexKV_Get(b *testing.B) {
	kv := NewRWMutex[string, string]()
	ctx := context.Background()

	// Pre-populate
	for i := 0; i < benchKeyCount; i++ {
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
}

// BenchmarkRWMutexKV_Set benchmarks Set under parallel access.
func BenchmarkRWMutexKV_Set(b *testing.B) {
	kv := NewRWMutex[string, string]()
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = kv.Set(ctx, strconv.Itoa(i%benchKeyCount), "value")
			i++
		}
	})
}

// BenchmarkRWMutexKV_Mixed benchmarks 80% read / 20% write under parallel access.
func BenchmarkRWMutexKV_Mixed(b *testing.B) {
	kv := NewRWMutex[string, string]()
	ctx := context.Background()

	// Pre-populate
	for i := 0; i < benchKeyCount; i++ {
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
}

// BenchmarkLRU_Get benchmarks LRU cache Get under parallel access.
func BenchmarkLRU_Get(b *testing.B) {
	kv, err := NewLRU[string, string](benchKeyCount, nil, 0)
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()

	// Pre-populate
	for i := 0; i < benchKeyCount; i++ {
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
}

// BenchmarkLRU_Set benchmarks LRU cache Set under parallel access.
func BenchmarkLRU_Set(b *testing.B) {
	kv, err := NewLRU[string, string](benchKeyCount, nil, 0)
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = kv.Set(ctx, strconv.Itoa(i%benchKeyCount), "value")
			i++
		}
	})
}

// BenchmarkLRU_Mixed benchmarks 80% read / 20% write under parallel access.
func BenchmarkLRU_Mixed(b *testing.B) {
	kv, err := NewLRU[string, string](benchKeyCount, nil, 0)
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()

	// Pre-populate
	for i := 0; i < benchKeyCount; i++ {
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
}
