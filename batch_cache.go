package db

import "context"

type batchCacheKV[K comparable, V any] struct {
	cache  *cacheKV[K, V]
	source BatchKV[K, V]
}

func NewBatchCacheKV[K comparable, V any](source BatchKV[K, V], options ...CacheOption[K, V]) *batchCacheKV[K, V] {
	cache := NewCacheKV(options...)
	return &batchCacheKV[K, V]{
		cache:  cache,
		source: source,
	}
}

func (c *batchCacheKV[K, V]) Get(ctx context.Context, keys []K) (result map[K]V, err error) {
	result = make(map[K]V, len(keys))
	var misses []K
	// todo: optimize cache.Get to lock/unlock once
	for _, key := range keys {
		v, e := c.cache.Get(ctx, key)
		if e != nil {
			misses = append(misses, key)
			err = e
			continue
		}
		result[key] = v
	}
	if c.source == nil {
		if len(misses) == 0 {
			return result, NotFound
		}
		return result, nil
	}
	get, e := c.source.Get(ctx, misses)
	if e != nil {
		err = e
	}
	for k, v := range get {
		_ = c.cache.Set(ctx, k, v)
		result[k] = v
	}
	return result, err
}

func (c *batchCacheKV[K, V]) Set(ctx context.Context, m map[K]V) (err error) {
	for k, v := range m {
		if e := c.cache.Set(ctx, k, v); e != nil {
			err = e
		}
	}
	if e := c.source.Set(ctx, m); e != nil {
		err = e
	}
	return err
}

func (c *batchCacheKV[K, V]) Del(ctx context.Context, keys []K) (err error) {
	for _, key := range keys {
		if e := c.cache.Del(ctx, key); e != nil {
			err = e
		}
	}
	if e := c.source.Del(ctx, keys); e != nil {
		err = e
	}
	return err
}
