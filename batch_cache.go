package db

import "context"

// batchCacheKV is a struct that contains a cache and a source BatchKV.
// It is used to cache batch operations.
type batchCacheKV[K comparable, V any] struct {
	cache  *cacheKV[K, V]
	source BatchKV[K, V]
}

// NewBatchCacheKV creates a new batchCacheKV instance with the given source and options.
func NewBatchCacheKV[K comparable, V any](source BatchKV[K, V], options ...CacheOption[K, V]) *batchCacheKV[K, V] {
	cache := NewCacheKV(options...)
	return &batchCacheKV[K, V]{
		cache:  cache,
		source: source,
	}
}

// Get retrieves the values for the given keys from the cache.
// If a key is not found in the cache, it is retrieved from the source BatchKV.
func (c *batchCacheKV[K, V]) Get(ctx context.Context, keys []K) (map[K]V, error) {
	result := make(map[K]V, len(keys))
	var misses []K
	for _, key := range keys {
		v, e := c.cache.Get(ctx, key)
		if e != nil {
			misses = append(misses, key)
			continue
		}
		result[key] = v
	}

	if c.source == nil {
		return result, nil
	}

	get, err := c.source.Get(ctx, misses)
	if err != nil {
		return result, err
	}
	for k, v := range get {
		_ = c.cache.Set(ctx, k, v)
		result[k] = v
	}
	return result, nil
}

// Set sets the values for the given keys in the cache.
// It also sets the values in the source BatchKV if it exists.
func (c *batchCacheKV[K, V]) Set(ctx context.Context, m map[K]V) error {
	for k, v := range m {
		_ = c.cache.Set(ctx, k, v)
	}
	if c.source == nil {
		return nil
	}
	return c.source.Set(ctx, m)
}

// Del deletes the values for the given keys from the cache.
// It also deletes the values from the source BatchKV if it exists.
func (c *batchCacheKV[K, V]) Del(ctx context.Context, keys []K) error {
	for _, key := range keys {
		_ = c.cache.Del(ctx, key)
	}
	if c.source == nil {
		return nil
	}
	return c.source.Del(ctx, keys)
}
