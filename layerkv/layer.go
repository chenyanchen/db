package layerkv

import (
	"context"
	"errors"
	"time"

	"github.com/chenyanchen/db"
)

type State int8

const (
	StateHit = iota
	StateMiss
	StateError
)

func (s State) String() string {
	switch s {
	case StateHit:
		return "hit"
	case StateMiss:
		return "miss"
	case StateError:
		return "error"
	}
	return "unknown"
}

type telemetryFn[K comparable, V any] func(key K, state State, layer int)

type layerKV[K comparable, V any] struct {
	layers []db.KV[K, V]

	// telFn is a telemetry function to record the state of the key.
	telFn telemetryFn[K, V]

	// writebackTimeout is the timeout for write back operation.
	// If it is zero, the write back operation will be synchronous,
	// Otherwise, it will be asynchronous with the given timeout.
	writebackTimeout time.Duration
}

func NewLayerKV[K comparable, V any](telFn telemetryFn[K, V], writebackTimeout time.Duration, layers ...db.KV[K, V]) *layerKV[K, V] {
	return &layerKV[K, V]{
		layers:           layers,
		telFn:            telFn,
		writebackTimeout: writebackTimeout,
	}
}

func (l *layerKV[K, V]) Get(ctx context.Context, k K) (V, error) {
	misses := make([]db.KV[K, V], 0, len(l.layers))
	for i, layer := range l.layers {
		got, err := layer.Get(ctx, k)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				l.telemetry(k, StateMiss, i)
				continue
			}

			l.telemetry(k, StateError, i)
			return got, err
		}

		// write back into missed layers.
		l.writeback(ctx, misses, k, got)

		l.telemetry(k, StateHit, i)
		return got, nil
	}

	var v V
	return v, db.ErrNotFound
}

func (l *layerKV[K, V]) telemetry(k K, state State, layer int) {
	if l.telFn != nil {
		// Record the state of the key.
		// layer+1 for human-readable.
		l.telFn(k, state, layer+1)
	}
}

func (l *layerKV[K, V]) writeback(ctx context.Context, layers []db.KV[K, V], k K, got V) {
	if l.writebackTimeout <= 0 {
		l._writeback(ctx, layers, k, got)
		return
	}

	go func() {
		asyncCtx, cancel := context.WithTimeout(ctx, l.writebackTimeout)
		defer cancel()
		l._writeback(asyncCtx, layers, k, got)
	}()
}

func (l *layerKV[K, V]) _writeback(ctx context.Context, layers []db.KV[K, V], k K, got V) {
	for _, layer := range layers {
		_ = layer.Set(ctx, k, got)
	}
}

func (l *layerKV[K, V]) Set(ctx context.Context, k K, v V) error {
	for _, layer := range l.layers {
		if err := layer.Set(ctx, k, v); err != nil {
			return err
		}
	}
	return nil
}

func (l *layerKV[K, V]) Del(ctx context.Context, k K) error {
	for _, layer := range l.layers {
		if err := layer.Del(ctx, k); err != nil {
			return err
		}
	}
	return nil
}
