package cachekv

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/Code-Hex/go-generics-cache/policy/fifo"
	"github.com/Code-Hex/go-generics-cache/policy/lfu"
	"github.com/Code-Hex/go-generics-cache/policy/lru"
	"github.com/Code-Hex/go-generics-cache/policy/mru"

	"github.com/chenyanchen/db"
)

type CacheOption[K comparable, V any] func(kv *cacheKV[K, V])

// WithTTL returns a CacheOption to set ttl function.
func WithTTL[K comparable, V any](ttlFn func(K) time.Duration) CacheOption[K, V] {
	return func(kv *cacheKV[K, V]) { kv.ttlFn = ttlFn }
}

// WithExpires returns a CacheOption to set solid ttl duration.
func WithExpires[K comparable, V any](ttl time.Duration) CacheOption[K, V] {
	return WithTTL[K, V](func(K) time.Duration { return ttl })
}

// WithSmoothExpires returns a CacheOption to set smooth ttl.
// The real TTL is a random value between [0.5*ttl, 1.5*ttl).
func WithSmoothExpires[K comparable, V any](ttl time.Duration) CacheOption[K, V] {
	return WithTTL[K, V](func(k K) time.Duration {
		return time.Duration((0.5 + rand.Float64()) * float64(ttl))
	})
}

// WithSource returns a CacheOption to set source KV-storage.
func WithSource[K comparable, V any](source db.KV[K, V]) CacheOption[K, V] {
	return func(kv *cacheKV[K, V]) { kv.source = source }
}

// WithTelemetryFunc returns a CacheOption to set telemetry prefix.
func WithTelemetryFunc[K comparable, V any](keyFn func(K, string)) CacheOption[K, V] {
	return func(kv *cacheKV[K, V]) { kv.telFn = keyFn }
}

// AsLRU returns a CacheOption to set LRU cache.
func AsLRU[K comparable, V any](capacity int) CacheOption[K, V] {
	return func(kv *cacheKV[K, V]) {
		kv.cache = cache.New[K, V](cache.AsLRU[K, V](lru.WithCapacity(capacity)))
	}
}

// AsLFU returns a CacheOption to set LFU cache.
func AsLFU[K comparable, V any](capacity int) CacheOption[K, V] {
	return func(kv *cacheKV[K, V]) {
		kv.cache = cache.New[K, V](cache.AsLFU[K, V](lfu.WithCapacity(capacity)))
	}
}

// AsFILO returns a CacheOption to set FILO cache.
func AsFILO[K comparable, V any](capacity int) CacheOption[K, V] {
	return func(kv *cacheKV[K, V]) {
		kv.cache = cache.New[K, V](cache.AsFIFO[K, V](fifo.WithCapacity(capacity)))
	}
}

// AsMRU returns a CacheOption to set MRU cache.
func AsMRU[K comparable, V any](capacity int) CacheOption[K, V] {
	return func(kv *cacheKV[K, V]) {
		kv.cache = cache.New[K, V](cache.AsMRU[K, V](mru.WithCapacity(capacity)))
	}
}

// cacheKV is a KV-storage with cache.
type cacheKV[K comparable, V any] struct {
	// cache
	cache *cache.Cache[K, V]

	// source KV-storage
	source db.KV[K, V]

	// ttl function returns ttl duration.
	ttlFn func(K) time.Duration

	// telemetry function
	telFn func(k K, scene string)
}

func New[K comparable, V any](options ...CacheOption[K, V]) *cacheKV[K, V] {
	kv := &cacheKV[K, V]{cache: cache.New[K, V]()}
	for _, opt := range options {
		opt(kv)
	}
	return kv
}

func (c *cacheKV[K, V]) Get(ctx context.Context, k K) (V, error) {
	v, ok := c.cache.Get(k)
	if ok {
		c.telemetry(k, "hit_mem")
		return v, nil
	}
	if c.source == nil {
		c.telemetry(k, "miss_mem")
		return v, fmt.Errorf("not found: %+v", k)
	}
	got, err := c.source.Get(ctx, k)
	if err != nil {
		c.telemetry(k, "miss_src")
		return got, fmt.Errorf("get from source: %w", err)
	}
	c.cache.Set(k, got, c.cacheOptions(k)...)
	c.telemetry(k, "hit_src")
	return got, nil
}

func (c *cacheKV[K, V]) Set(ctx context.Context, k K, v V) error {
	c.cache.Set(k, v, c.cacheOptions(k)...)
	if c.source != nil {
		return c.source.Set(ctx, k, v)
	}
	return nil
}

func (c *cacheKV[K, V]) cacheOptions(k K) []cache.ItemOption {
	var opts []cache.ItemOption
	if c.ttlFn != nil {
		opts = append(opts, cache.WithExpiration(c.ttlFn(k)))
	}
	return opts
}

func (c *cacheKV[K, V]) Del(ctx context.Context, k K) error {
	c.cache.Delete(k)
	if c.source != nil {
		return c.source.Del(ctx, k)
	}
	return nil
}

func (c *cacheKV[K, V]) telemetry(k K, scene string) {
	if c.telFn == nil {
		return
	}
	c.telFn(k, scene)
}
