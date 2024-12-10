package layerkv

import (
	"context"
	"errors"
	"maps"
	"slices"

	"github.com/chenyanchen/db"
)

type batch[K comparable, V any] struct {
	cache db.BatchKV[K, V]
	store db.BatchKV[K, V]
}

func NewBatch[K comparable, V any](cache, store db.BatchKV[K, V]) (*batch[K, V], error) {
	if cache == nil {
		return nil, errors.New("cache is nil")
	}
	if store == nil {
		return nil, errors.New("store is nil")
	}
	return &batch[K, V]{
		cache: cache,
		store: store,
	}, nil
}

func (l batch[K, V]) Get(ctx context.Context, keys []K) (map[K]V, error) {
	cache, err := l.cache.Get(ctx, keys)
	if err != nil {
		return nil, err
	}

	if len(cache) == len(keys) {
		return cache, nil
	}

	miss := make([]K, 0, len(keys)-len(cache))
	for _, key := range keys {
		if _, ok := cache[key]; !ok {
			miss = append(miss, key)
		}
	}

	store, err := l.store.Get(ctx, miss)
	if err != nil {
		return nil, err
	}

	for k, v := range store {
		cache[k] = v
	}

	return cache, l.cache.Set(ctx, store)
}

func (l batch[K, V]) Set(ctx context.Context, kvs map[K]V) error {
	if err := l.store.Set(ctx, kvs); err != nil {
		return err
	}

	return l.cache.Del(ctx, slices.Collect(maps.Keys(kvs)))
}

func (l batch[K, V]) Del(ctx context.Context, keys []K) error {
	if err := l.store.Del(ctx, keys); err != nil {
		return err
	}

	return l.cache.Del(ctx, keys)
}
