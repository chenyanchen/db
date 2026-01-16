package layerkv

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	kv "github.com/chenyanchen/kv"
	"github.com/chenyanchen/kv/mocks"
)

func Test_layerKV_Get(t *testing.T) {
	type args[K comparable] struct {
		ctx context.Context
		k   K
	}
	type testCase[K comparable, V any] struct {
		name    string
		l       layerKV[K, V]
		args    args[K]
		want    V
		wantErr assert.ErrorAssertionFunc
	}
	tests := []testCase[string, string]{
		{
			name: "from cache",
			l: layerKV[string, string]{
				cache: &mocks.MockKVStore[string, string]{
					GetFunc: func(ctx context.Context, k string) (string, error) {
						return "value", nil
					},
				},
			},
			args:    args[string]{context.Background(), "key"},
			want:    "value",
			wantErr: assert.NoError,
		}, {
			name: "cache error",
			l: layerKV[string, string]{
				cache: &mocks.MockKVStore[string, string]{
					GetFunc: func(ctx context.Context, k string) (string, error) {
						return "", assert.AnError
					},
				},
			},
			args:    args[string]{context.Background(), "key"},
			want:    "",
			wantErr: assert.Error,
		}, {
			name: "store error",
			l: layerKV[string, string]{
				cache: &mocks.MockKVStore[string, string]{
					GetFunc: func(ctx context.Context, k string) (string, error) {
						return "", kv.ErrNotFound
					},
				},
				store: &mocks.MockKVStore[string, string]{
					GetFunc: func(ctx context.Context, k string) (string, error) {
						return "", assert.AnError
					},
				},
			},
			args:    args[string]{context.Background(), "key"},
			want:    "",
			wantErr: assert.Error,
		}, {
			name: "cache set error",
			l: layerKV[string, string]{
				cache: &mocks.MockKVStore[string, string]{
					GetFunc: func(ctx context.Context, k string) (string, error) {
						return "", kv.ErrNotFound
					},
					SetFunc: func(ctx context.Context, k string, v string) error {
						return assert.AnError
					},
				},
				store: &mocks.MockKVStore[string, string]{
					GetFunc: func(ctx context.Context, k string) (string, error) {
						return "value", nil
					},
				},
			},
			args:    args[string]{context.Background(), "key"},
			want:    "value",
			wantErr: assert.Error,
		}, {
			name: "no error",
			l: layerKV[string, string]{
				cache: &mocks.MockKVStore[string, string]{
					GetFunc: func(ctx context.Context, k string) (string, error) {
						return "", kv.ErrNotFound
					},
					SetFunc: func(ctx context.Context, k string, v string) error {
						return nil
					},
				},
				store: &mocks.MockKVStore[string, string]{
					GetFunc: func(ctx context.Context, k string) (string, error) {
						return "value", nil
					},
				},
			},
			args:    args[string]{context.Background(), "key"},
			want:    "value",
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.l.Get(tt.args.ctx, tt.args.k)
			if !tt.wantErr(t, err, fmt.Sprintf("Get(%v, %v)", tt.args.ctx, tt.args.k)) {
				return
			}
			assert.Equalf(t, tt.want, got, "Get(%v, %v)", tt.args.ctx, tt.args.k)
		})
	}
}

func Test_layerKV_Set(t *testing.T) {
	type args[K comparable, V any] struct {
		ctx context.Context
		k   K
		v   V
	}
	type testCase[K comparable, V any] struct {
		name    string
		l       layerKV[K, V]
		args    args[K, V]
		wantErr assert.ErrorAssertionFunc
	}
	tests := []testCase[string, string]{
		{
			name: "store error",
			l: layerKV[string, string]{
				store: &mocks.MockKVStore[string, string]{
					SetFunc: func(ctx context.Context, k string, v string) error {
						return assert.AnError
					},
				},
			},
			args:    args[string, string]{context.Background(), "key", "value"},
			wantErr: assert.Error,
		}, {
			name: "cache error",
			l: layerKV[string, string]{
				store: &mocks.MockKVStore[string, string]{
					SetFunc: func(ctx context.Context, k string, v string) error {
						return nil
					},
				},
				cache: &mocks.MockKVStore[string, string]{
					DelFunc: func(ctx context.Context, k string) error {
						return assert.AnError
					},
				},
			},
			args:    args[string, string]{context.Background(), "key", "value"},
			wantErr: assert.Error,
		}, {
			name: "no error",
			l: layerKV[string, string]{
				cache: &mocks.MockKVStore[string, string]{
					DelFunc: func(ctx context.Context, k string) error {
						return nil
					},
				},
				store: &mocks.MockKVStore[string, string]{
					SetFunc: func(ctx context.Context, k string, v string) error {
						return nil
					},
				},
			},
			args:    args[string, string]{context.Background(), "key", "value"},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, tt.l.Set(tt.args.ctx, tt.args.k, tt.args.v), fmt.Sprintf("Set(%v, %v, %v)", tt.args.ctx, tt.args.k, tt.args.v))
		})
	}
}

func Test_layerKV_Del(t *testing.T) {
	type args[K comparable] struct {
		ctx context.Context
		k   K
	}
	type testCase[K comparable, V any] struct {
		name    string
		l       layerKV[K, V]
		args    args[K]
		wantErr assert.ErrorAssertionFunc
	}
	tests := []testCase[string, string]{
		{
			name: "store error",
			l: layerKV[string, string]{
				store: &mocks.MockKVStore[string, string]{
					DelFunc: func(ctx context.Context, k string) error {
						return assert.AnError
					},
				},
			},
			args:    args[string]{context.Background(), "key"},
			wantErr: assert.Error,
		}, {
			name: "cache error",
			l: layerKV[string, string]{
				store: &mocks.MockKVStore[string, string]{
					DelFunc: func(ctx context.Context, k string) error {
						return nil
					},
				},
				cache: &mocks.MockKVStore[string, string]{
					DelFunc: func(ctx context.Context, k string) error {
						return assert.AnError
					},
				},
			},
			args:    args[string]{context.Background(), "key"},
			wantErr: assert.Error,
		}, {
			name: "no error",
			l: layerKV[string, string]{
				cache: &mocks.MockKVStore[string, string]{
					DelFunc: func(ctx context.Context, k string) error {
						return nil
					},
				},
				store: &mocks.MockKVStore[string, string]{
					DelFunc: func(ctx context.Context, k string) error {
						return nil
					},
				},
			},
			args:    args[string]{context.Background(), "key"},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, tt.l.Del(tt.args.ctx, tt.args.k), fmt.Sprintf("Del(%v, %v)", tt.args.ctx, tt.args.k))
		})
	}
}
