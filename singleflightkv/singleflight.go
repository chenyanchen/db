package singleflightkv

import (
	"context"
	"errors"

	"github.com/chenyanchen/sync/singleflight"

	"github.com/chenyanchen/db"
)

// sfKV represents a single-flight KV-storage to avoid concurrent
// operations on the same key with source KV. In other words, it ensures
// that only one operation is performed on the same key at the same time.
type sfKV[K comparable, V any] struct {
	// source KV-storage
	source db.KV[K, V]

	// single-flight gourp.
	group singleflight.Group[K, V]
}

func New[K comparable, V any](source db.KV[K, V]) (*sfKV[K, V], error) {
	if source == nil {
		return nil, errors.New("source KV-storage is required")
	}
	return &sfKV[K, V]{source: source}, nil
}

func (s *sfKV[K, V]) Get(ctx context.Context, k K) (V, error) {
	v, err, _ := s.group.Do(k, func() (V, error) { return s.source.Get(ctx, k) })
	return v, err
}

func (s *sfKV[K, V]) Set(ctx context.Context, k K, v V) error {
	_, err, _ := s.group.Do(k, func() (V, error) { return v, s.source.Set(ctx, k, v) })
	return err
}

func (s *sfKV[K, V]) Del(ctx context.Context, k K) error {
	var v V
	_, err, _ := s.group.Do(k, func() (V, error) { return v, s.source.Del(ctx, k) })
	return err
}
