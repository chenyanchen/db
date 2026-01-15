package cachekv

import (
	"context"
	"hash/maphash"
	"unsafe"

	"github.com/chenyanchen/db"
)

const defaultShardCount = 32

type shardedKV[K comparable, V any] struct {
	shards []*rwMutexKV[K, V]
	seed   maphash.Seed
	mask   uint64
}

// NewSharded creates a sharded KV store with numShards partitions.
// numShards should be a power of 2 for efficient modulo via bitmask.
// If numShards <= 0, defaults to 32.
// Sharding reduces lock contention under high concurrent access.
func NewSharded[K comparable, V any](numShards int) *shardedKV[K, V] {
	if numShards <= 0 {
		numShards = defaultShardCount
	}
	// Round up to next power of 2
	numShards = nextPowerOf2(numShards)

	shards := make([]*rwMutexKV[K, V], numShards)
	for i := range shards {
		shards[i] = NewRWMutex[K, V]()
	}

	return &shardedKV[K, V]{
		shards: shards,
		seed:   maphash.MakeSeed(),
		mask:   uint64(numShards - 1),
	}
}

func nextPowerOf2(n int) int {
	if n <= 1 {
		return 1
	}
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++
	return n
}

func (s *shardedKV[K, V]) getShard(k K) *rwMutexKV[K, V] {
	var h maphash.Hash
	h.SetSeed(s.seed)

	// Hash based on key type for optimal performance
	switch key := any(k).(type) {
	case string:
		h.WriteString(key)
	case int:
		b := (*[unsafe.Sizeof(key)]byte)(unsafe.Pointer(&key))
		h.Write(b[:])
	case int64:
		b := (*[unsafe.Sizeof(key)]byte)(unsafe.Pointer(&key))
		h.Write(b[:])
	case int32:
		b := (*[unsafe.Sizeof(key)]byte)(unsafe.Pointer(&key))
		h.Write(b[:])
	case uint:
		b := (*[unsafe.Sizeof(key)]byte)(unsafe.Pointer(&key))
		h.Write(b[:])
	case uint64:
		b := (*[unsafe.Sizeof(key)]byte)(unsafe.Pointer(&key))
		h.Write(b[:])
	case uint32:
		b := (*[unsafe.Sizeof(key)]byte)(unsafe.Pointer(&key))
		h.Write(b[:])
	default:
		// Fallback: use the memory address of the key as a poor hash
		// This works but has worse distribution for non-pointer types
		ptr := unsafe.Pointer(&k)
		b := (*[unsafe.Sizeof(ptr)]byte)(unsafe.Pointer(&ptr))
		h.Write(b[:])
	}

	idx := h.Sum64() & s.mask
	return s.shards[idx]
}

func (s *shardedKV[K, V]) Get(ctx context.Context, k K) (V, error) {
	shard := s.getShard(k)
	shard.mu.RLock()
	v, ok := shard.m[k]
	shard.mu.RUnlock()

	if !ok {
		return v, db.ErrNotFound
	}
	return v, nil
}

func (s *shardedKV[K, V]) Set(ctx context.Context, k K, v V) error {
	shard := s.getShard(k)
	shard.mu.Lock()
	shard.m[k] = v
	shard.mu.Unlock()
	return nil
}

func (s *shardedKV[K, V]) Del(ctx context.Context, k K) error {
	shard := s.getShard(k)
	shard.mu.Lock()
	delete(shard.m, k)
	shard.mu.Unlock()
	return nil
}

// Len returns the total number of items across all shards.
func (s *shardedKV[K, V]) Len() int {
	var count int
	for _, shard := range s.shards {
		shard.mu.RLock()
		count += len(shard.m)
		shard.mu.RUnlock()
	}
	return count
}
