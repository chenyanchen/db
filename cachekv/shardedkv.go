package cachekv

import (
	"context"
	"hash/maphash"

	kv "github.com/chenyanchen/kv"
)

const defaultShardCount = 32

type shardedKV[K comparable, V any] struct {
	shards []*rwMutexKV[K, V]
	seed   maphash.Seed
}

// NewSharded creates a sharded KV store with numShards partitions.
// If numShards <= 0, defaults to 32.
// Sharding reduces lock contention under high concurrent access.
func NewSharded[K comparable, V any](numShards int) *shardedKV[K, V] {
	if numShards <= 0 {
		numShards = defaultShardCount
	}

	shards := make([]*rwMutexKV[K, V], numShards)
	for i := range shards {
		shards[i] = NewRWMutex[K, V]()
	}

	return &shardedKV[K, V]{
		shards: shards,
		seed:   maphash.MakeSeed(),
	}
}

func (s *shardedKV[K, V]) getShard(k K) *rwMutexKV[K, V] {
	return s.shards[maphash.Comparable(s.seed, k)%uint64(len(s.shards))]
}

func (s *shardedKV[K, V]) Get(ctx context.Context, k K) (V, error) {
	shard := s.getShard(k)
	shard.mu.RLock()
	v, ok := shard.m[k]
	shard.mu.RUnlock()

	if !ok {
		return v, kv.ErrNotFound
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
