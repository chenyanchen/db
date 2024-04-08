package cachekv

import (
	"context"
	"time"

	"github.com/hashicorp/golang-lru/v2"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/hashicorp/golang-lru/v2/simplelru"

	"github.com/chenyanchen/db"
)

type lruKV[K comparable, V any] struct {
	cache simplelru.LRUCache[K, V]
}

func NewLRU[K comparable, V any](size int, onEvict func(K, V), ttl time.Duration) (*lruKV[K, V], error) {
	var cache simplelru.LRUCache[K, V]
	var err error

	if ttl > 0 {
		cache = expirable.NewLRU[K, V](size, onEvict, ttl)
	} else {
		cache, err = lru.NewWithEvict[K, V](size, onEvict)
	}

	return &lruKV[K, V]{cache: cache}, err
}

func (c *lruKV[K, V]) Get(ctx context.Context, k K) (V, error) {
	v, ok := c.cache.Get(k)
	if ok {
		return v, nil
	}
	return v, db.ErrNotFound
}

func (c *lruKV[K, V]) Set(ctx context.Context, k K, v V) error {
	c.cache.Add(k, v)
	return nil
}

func (c *lruKV[K, V]) Del(ctx context.Context, k K) error {
	c.cache.Remove(k)
	return nil
}
