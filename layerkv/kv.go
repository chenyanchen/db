package layerkv

import (
	"context"
	"errors"

	"github.com/chenyanchen/db"
)

type layerKV[K comparable, V any] struct {
	cache db.KV[K, V]
	store db.KV[K, V]
}

func New[K comparable, V any](cache, store db.KV[K, V]) (*layerKV[K, V], error) {
	if cache == nil {
		return nil, errors.New("cache is nil")
	}
	if store == nil {
		return nil, errors.New("store is nil")
	}
	return &layerKV[K, V]{cache: cache, store: store}, nil
}

func (l *layerKV[K, V]) Get(ctx context.Context, k K) (V, error) {
	v, err := l.cache.Get(ctx, k)
	if err == nil {
		return v, nil
	}

	if !errors.Is(err, db.ErrNotFound) {
		return v, err
	}

	v, err = l.store.Get(ctx, k)
	if err != nil {
		return v, err
	}

	return v, l.cache.Set(ctx, k, v)
}

func (l *layerKV[K, V]) Set(ctx context.Context, k K, v V) error {
	if err := l.store.Set(ctx, k, v); err != nil {
		return err
	}

	return l.cache.Del(ctx, k)
}

func (l *layerKV[K, V]) Del(ctx context.Context, k K) error {
	if err := l.store.Del(ctx, k); err != nil {
		return err
	}

	return l.cache.Del(ctx, k)
}
