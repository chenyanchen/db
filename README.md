# db

Generic key-value storage abstraction library for Go with composable implementations.

## Features

- Generic interfaces using Go generics (`KV[K, V]` and `BatchKV[K, V]`)
- Composable/stackable implementations
- Context-aware operations
- Multiple cache implementations with different performance characteristics

## Installation

```bash
go get github.com/chenyanchen/kv
```

## Implementations

### cachekv

| Implementation               | Description                      | Use Case                    |
| ---------------------------- | -------------------------------- | --------------------------- |
| `NewRWMutex()`               | Simple RWMutex-protected map     | Low concurrency workloads   |
| `NewSharded(numShards)`      | Sharded map with per-shard locks | High concurrency workloads  |
| `NewLRU(size, onEvict, ttl)` | LRU cache with optional TTL      | Bounded cache with eviction |

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

### Sharded vs RWMutex KV (10 CPU cores, parallel access, Apple M4)

```
BenchmarkRWMutexKV_Get-10              12820437    91.37 ns/op
BenchmarkShardedKV_Get/shards=16-10    43370292    29.54 ns/op    (3.1x faster)
BenchmarkShardedKV_Get/shards=32-10    65453353    18.97 ns/op    (4.8x faster)
BenchmarkShardedKV_Get/shards=64-10    80310759    13.96 ns/op    (6.5x faster)

BenchmarkRWMutexKV_Set-10              9575170     115.1 ns/op
BenchmarkShardedKV_Set/shards=16-10    25261207    44.10 ns/op    (2.6x faster)
BenchmarkShardedKV_Set/shards=32-10    32026902    34.11 ns/op    (3.4x faster)
BenchmarkShardedKV_Set/shards=64-10    35371189    31.12 ns/op    (3.7x faster)

BenchmarkRWMutexKV_Mixed-10              21981846    56.47 ns/op
BenchmarkShardedKV_Mixed/shards=16-10    32515441    35.50 ns/op    (1.6x faster)
BenchmarkShardedKV_Mixed/shards=32-10    41161084    26.79 ns/op    (2.1x faster)
BenchmarkShardedKV_Mixed/shards=64-10    50309277    21.81 ns/op    (2.6x faster)
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
    "github.com/chenyanchen/kv/cachekv"
    "github.com/chenyanchen/kv/layerkv"
)

func main() {
    ctx := context.Background()

    // Simple sharded cache for high concurrency
    cache := cachekv.NewSharded[string, string](32)
    cache.Set(ctx, "key", "value")
    v, _ := cache.Get(ctx, "key")

    // Layered cache with database backend
    lru, _ := cachekv.NewLRU[string, string](1000, nil, 0)
    store := &myDatabaseKV{} // implements kv.KV[string, string]
    layered, _ := layerkv.New(lru, store, layerkv.WithWriteThrough())
    layered.Get(ctx, "key") // checks cache first, then store
}
```

## License

MIT
