# db

Generic key-value storage abstraction library for Go with composable implementations.

## Features

- Generic interfaces using Go generics (`KV[K, V]` and `BatchKV[K, V]`)
- Composable/stackable implementations
- Context-aware operations
- Multiple cache implementations with different performance characteristics

## Installation

```bash
go get github.com/chenyanchen/db
```

## Implementations

### cachekv

| Implementation | Description | Use Case |
|----------------|-------------|----------|
| `NewRWMutex()` | Simple RWMutex-protected map | Low concurrency workloads |
| `NewSharded(numShards)` | Sharded map with per-shard locks | High concurrency workloads |
| `NewLRU(size, onEvict, ttl)` | LRU cache with optional TTL | Bounded cache with eviction |

### layerkv

Composes cache and store layers with configurable write strategies:

```go
// Default: Write-invalidate (deletes from cache on Set)
kv, _ := layerkv.New(cache, store)

// Write-through (updates cache on Set)
kv, _ := layerkv.New(cache, store, layerkv.WithWriteThrough())
```

### singleflightkv

Deduplicates concurrent requests for the same key.

## Benchmarks

### Sharded vs RWMutex KV (10 CPU cores, parallel access)

```
BenchmarkRWMutexKV_Get-10           13311891       92 ns/op
BenchmarkShardedKV_Get/shards=32-10 60517676       19 ns/op   (4.7x faster)
BenchmarkShardedKV_Get/shards=64-10 79021450       15 ns/op   (6x faster)

BenchmarkRWMutexKV_Set-10           10492590      115 ns/op
BenchmarkShardedKV_Set/shards=32-10 29543521       41 ns/op   (2.8x faster)
BenchmarkShardedKV_Set/shards=64-10 32410058       36 ns/op   (3.2x faster)

BenchmarkRWMutexKV_Mixed-10         21990154       55 ns/op
BenchmarkShardedKV_Mixed/shards=32-10 38894740     30 ns/op   (1.8x faster)
BenchmarkShardedKV_Mixed/shards=64-10 48160051     26 ns/op   (2.1x faster)
```

### Write Strategy Comparison

```
BenchmarkLayerKV_SetThenGet_WriteInvalidate-10  25039428   48 ns/op
BenchmarkLayerKV_SetThenGet_WriteThrough-10     40282930   30 ns/op  (1.6x faster)
```

## Usage

```go
package main

import (
    "context"
    "github.com/chenyanchen/db/cachekv"
    "github.com/chenyanchen/db/layerkv"
)

func main() {
    ctx := context.Background()

    // Simple sharded cache for high concurrency
    cache := cachekv.NewSharded[string, string](32)
    cache.Set(ctx, "key", "value")
    v, _ := cache.Get(ctx, "key")

    // Layered cache with database backend
    lru, _ := cachekv.NewLRU[string, string](1000, nil, 0)
    store := &myDatabaseKV{} // implements db.KV[string, string]
    layered, _ := layerkv.New(lru, store, layerkv.WithWriteThrough())
    layered.Get(ctx, "key") // checks cache first, then store
}
```

## License

MIT
