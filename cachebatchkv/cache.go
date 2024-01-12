package cachebatchkv

import (
	"context"
	"errors"
	"fmt"

	"github.com/chenyanchen/db"
	"github.com/chenyanchen/db/cachekv"
)

// cacheBatchKV is a struct that contains a cache and a source BatchKV.
// It is used to cache batch operations.
//
// Important: cacheBatchKV are not guaranteed to get all the values of keys, it guarantees no error.
type cacheBatchKV[K comparable, V any] struct {
	cache db.KV[K, V]

	source db.BatchKV[K, V]
}

// New creates a new cacheBatchKV instance with the given source and options.
func New[K comparable, V any](
	source db.BatchKV[K, V],
	options ...cachekv.Option[K, V],
) *cacheBatchKV[K, V] {
	return &cacheBatchKV[K, V]{
		cache:  cachekv.New(options...),
		source: source,
	}
}

// Get retrieves the values for the given keys from the cache.
// If a key is not found in the cache, it is retrieved from the source BatchKV.
func (c *cacheBatchKV[K, V]) Get(ctx context.Context, keys []K) (map[K]V, error) {
	result := make(map[K]V, len(keys))

	// misses is a slice that contains the keys that are not found in the cache.
	var misses []K

	for _, key := range keys {
		v, err := c.cache.Get(ctx, key)
		if err == nil {
			result[key] = v
			continue
		}

		if errors.Is(err, db.ErrNotFound) {
			misses = append(misses, key)
			continue
		}

		return nil, err
	}

	if c.source == nil {
		return result, nil
	}

	get, err := c.source.Get(ctx, misses)
	if err != nil {
		return result, err
	}

	for k, v := range get {
		result[k] = v
		if err = c.cache.Set(ctx, k, v); err != nil {
			return nil, fmt.Errorf("set cache: %w", err)
		}
	}

	return result, nil
}

// Set sets the values for the given keys in the cache.
// It also sets the values in the source BatchKV if it exists.
func (c *cacheBatchKV[K, V]) Set(ctx context.Context, m map[K]V) error {
	for k, v := range m {
		if err := c.cache.Set(ctx, k, v); err != nil {
			return fmt.Errorf("set cache: %w", err)
		}
	}

	if c.source == nil {
		return nil
	}

	return c.source.Set(ctx, m)
}

// Del deletes the values for the given keys from the cache.
// It also deletes the values from the source BatchKV if it exists.
func (c *cacheBatchKV[K, V]) Del(ctx context.Context, keys []K) error {
	for _, key := range keys {
		if err := c.cache.Del(ctx, key); err != nil {
			return fmt.Errorf("del cache: %w", err)
		}
	}

	if c.source == nil {
		return nil
	}

	return c.source.Del(ctx, keys)
}
