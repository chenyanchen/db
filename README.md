# kv

Generic key-value storage abstraction library for Go with composable implementations.

## Installation

```bash
go get github.com/chenyanchen/kv
```

## Core Interfaces

The library provides two generic interfaces for key-value operations:

```go
// KV represents a key-value storage for single-key operations.
type KV[K comparable, V any] interface {
    Get(ctx context.Context, k K) (V, error)
    Set(ctx context.Context, k K, v V) error
    Del(ctx context.Context, k K) error
}

// BatchKV represents a key-value storage for batch operations.
type BatchKV[K comparable, V any] interface {
    Get(ctx context.Context, keys []K) (map[K]V, error)
    Set(ctx context.Context, kvs map[K]V) error
    Del(ctx context.Context, keys []K) error
}

// ErrNotFound is returned when a key is not found.
var ErrNotFound = errors.New("not found")
```

## Implementing Your Own KV

Implement the `kv.KV` interface to integrate any storage backend:

```go
type databaseKV struct {
    db *sql.DB
}

func (s *databaseKV) Get(ctx context.Context, id int) (*User, error) {
    var user User
    err := s.db.QueryRowContext(ctx, "SELECT id, name FROM users WHERE id = ?", id).
        Scan(&user.ID, &user.Name)
    if errors.Is(err, sql.ErrNoRows) {
        return nil, kv.ErrNotFound
    }
    return &user, err
}

func (s *databaseKV) Set(ctx context.Context, id int, user *User) error {
    _, err := s.db.ExecContext(ctx,
        "INSERT INTO users (id, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE name = ?",
        id, user.Name, user.Name)
    return err
}

func (s *databaseKV) Del(ctx context.Context, id int) error {
    _, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE id = ?", id)
    return err
}
```

## Built-in Implementations

### cachekv

In-memory cache implementations:

| Implementation               | Description                      | Use Case                    |
| ---------------------------- | -------------------------------- | --------------------------- |
| `NewRWMutex()`               | Simple RWMutex-protected map     | Low concurrency workloads   |
| `NewSharded(numShards)`      | Sharded map with per-shard locks | High concurrency workloads  |
| `NewLRU(size, onEvict, ttl)` | LRU cache with optional TTL      | Bounded cache with eviction |

## Composition

The power of `kv.KV` comes from composing implementations together.

### layerkv - Cache + Store Layers

Compose a cache layer with a persistent store:

```go
cache, _ := cachekv.NewLRU[int, *User](1000, nil, 0)
store := &databaseKV{db: db}

// Cache-aside pattern: checks cache first, falls back to store
userKV, _ := layerkv.New(cache, store)

// With write-through: updates cache on Set instead of invalidating
userKV, _ := layerkv.New(cache, store, layerkv.WithWriteThrough())
```

### singleflightkv - Request Deduplication

Prevent duplicate concurrent requests for the same key:

```go
store := &databaseKV{db: db}
userKV, _ := singleflightkv.New(store)
```

### Custom Wrappers - Telemetry Example

Create your own wrapper to add cross-cutting concerns:

```go
type telemetry[K comparable, V any] struct {
    next   kv.KV[K, V]
    record func(operation string, duration time.Duration)
}

func (t telemetry[K, V]) Get(ctx context.Context, k K) (V, error) {
    start := time.Now()
    v, err := t.next.Get(ctx, k)
    t.record("Get", time.Since(start))
    return v, err
}

func (t telemetry[K, V]) Set(ctx context.Context, k K, v V) error {
    start := time.Now()
    err := t.next.Set(ctx, k, v)
    t.record("Set", time.Since(start))
    return err
}

func (t telemetry[K, V]) Del(ctx context.Context, k K) error {
    start := time.Now()
    err := t.next.Del(ctx, k)
    t.record("Del", time.Since(start))
    return err
}
```

### Full Composition Example

Combine multiple layers for a production-ready setup:

```go
// 1. Database backend (your implementation)
dbKV := &databaseKV{db: db}

// 2. Protect database with request deduplication
dbKV, _ = singleflightkv.New(dbKV)

// 3. Add telemetry to database
dbWithMetrics := NewTelemetry(dbKV, dbRecorder)

// 4. LRU cache layer
cache, _ := cachekv.NewLRU[int, *User](1000, nil, time.Minute*5)

// 5. Add telemetry to cache
cacheWithMetrics := NewTelemetry(cache, cacheRecorder)

// 6. Compose: cache + store with write-through
userKV, _ := layerkv.New(cacheWithMetrics, dbWithMetrics, layerkv.WithWriteThrough())
```

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

## License

MIT
