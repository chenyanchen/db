package layerkv

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/chenyanchen/kv/mocks"
)

func Test_batch_Get(t *testing.T) {
	type args[K comparable] struct {
		ctx  context.Context
		keys []K
	}
	type testCase[K comparable, V any] struct {
		name    string
		l       batch[K, V]
		args    args[K]
		want    map[K]V
		wantErr assert.ErrorAssertionFunc
	}
	tests := []testCase[string, string]{
		{
			name: "cache error",
			l: batch[string, string]{
				cache: &mocks.MockBatchKVStore[string, string]{
					GetFunc: func(ctx context.Context, keys []string) (map[string]string, error) {
						return nil, assert.AnError
					},
				},
			},
			args:    args[string]{context.Background(), []string{"key1", "key2"}},
			want:    nil,
			wantErr: assert.Error,
		}, {
			name: "all from cache",
			l: batch[string, string]{
				cache: &mocks.MockBatchKVStore[string, string]{
					GetFunc: func(ctx context.Context, keys []string) (map[string]string, error) {
						return map[string]string{
							"key1": "value1",
							"key2": "value2",
						}, nil
					},
				},
			},
			args: args[string]{context.Background(), []string{"key1", "key2"}},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			wantErr: assert.NoError,
		}, {
			name: "store error",
			l: batch[string, string]{
				cache: &mocks.MockBatchKVStore[string, string]{
					GetFunc: func(ctx context.Context, keys []string) (map[string]string, error) {
						return map[string]string{
							"key1": "value1",
						}, nil
					},
				},
				store: &mocks.MockBatchKVStore[string, string]{
					GetFunc: func(ctx context.Context, keys []string) (map[string]string, error) {
						return nil, assert.AnError
					},
				},
			},
			args:    args[string]{context.Background(), []string{"key1", "key2"}},
			want:    nil,
			wantErr: assert.Error,
		}, {
			name: "mixed",
			l: batch[string, string]{
				cache: &mocks.MockBatchKVStore[string, string]{
					GetFunc: func(ctx context.Context, keys []string) (map[string]string, error) {
						return map[string]string{"key1": "value1"}, nil
					},
					SetFunc: func(ctx context.Context, m map[string]string) error {
						return nil
					},
				},
				store: &mocks.MockBatchKVStore[string, string]{
					GetFunc: func(ctx context.Context, keys []string) (map[string]string, error) {
						return map[string]string{"key2": "value2"}, nil
					},
				},
			},
			args: args[string]{context.Background(), []string{"key1", "key2"}},
			want: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.l.Get(tt.args.ctx, tt.args.keys)
			if !tt.wantErr(t, err, fmt.Sprintf("Get(%v, %v)", tt.args.ctx, tt.args.keys)) {
				return
			}
			assert.Equalf(t, tt.want, got, "Get(%v, %v)", tt.args.ctx, tt.args.keys)
		})
	}
}

func Test_batch_Set(t *testing.T) {
	type args[K comparable, V any] struct {
		ctx context.Context
		kvs map[K]V
	}
	type testCase[K comparable, V any] struct {
		name    string
		l       batch[K, V]
		args    args[K, V]
		wantErr assert.ErrorAssertionFunc
	}
	tests := []testCase[string, string]{
		{
			name: "store error",
			l: batch[string, string]{
				store: &mocks.MockBatchKVStore[string, string]{
					SetFunc: func(ctx context.Context, m map[string]string) error {
						return assert.AnError
					},
				},
			},
			args:    args[string, string]{context.Background(), map[string]string{"key1": "value1"}},
			wantErr: assert.Error,
		}, {
			name: "cache error",
			l: batch[string, string]{
				store: &mocks.MockBatchKVStore[string, string]{
					SetFunc: func(ctx context.Context, m map[string]string) error {
						return nil
					},
				},
				cache: &mocks.MockBatchKVStore[string, string]{
					DelFunc: func(ctx context.Context, keys []string) error {
						return assert.AnError
					},
				},
			},
			args:    args[string, string]{context.Background(), map[string]string{"key1": "value1"}},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, tt.l.Set(tt.args.ctx, tt.args.kvs), fmt.Sprintf("Set(%v, %v)", tt.args.ctx, tt.args.kvs))
		})
	}
}

func Test_batch_Del(t *testing.T) {
	type args[K comparable] struct {
		ctx  context.Context
		keys []K
	}
	type testCase[K comparable, V any] struct {
		name    string
		l       batch[K, V]
		args    args[K]
		wantErr assert.ErrorAssertionFunc
	}
	tests := []testCase[string, string]{
		{
			name: "store error",
			l: batch[string, string]{
				store: &mocks.MockBatchKVStore[string, string]{
					DelFunc: func(ctx context.Context, keys []string) error {
						return assert.AnError
					},
				},
			},
			args:    args[string]{},
			wantErr: assert.Error,
		}, {
			name: "cache error",
			l: batch[string, string]{
				store: &mocks.MockBatchKVStore[string, string]{
					DelFunc: func(ctx context.Context, keys []string) error {
						return nil
					},
				},
				cache: &mocks.MockBatchKVStore[string, string]{
					DelFunc: func(ctx context.Context, keys []string) error {
						return assert.AnError
					},
				},
			},
			args:    args[string]{},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, tt.l.Del(tt.args.ctx, tt.args.keys), fmt.Sprintf("Del(%v, %v)", tt.args.ctx, tt.args.keys))
		})
	}
}
