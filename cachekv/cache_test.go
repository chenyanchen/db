package cachekv

import (
	"context"
	"errors"
	"fmt"
	"testing"

	cache "github.com/Code-Hex/go-generics-cache"
	"github.com/stretchr/testify/assert"

	"github.com/chenyanchen/db"
	"github.com/chenyanchen/db/mocks"
)

func Test_cacheKV_Get(t *testing.T) {
	type args[K comparable] struct {
		ctx context.Context
		k   K
	}
	type testCase[K comparable, V any] struct {
		name    string
		c       *cacheKV[K, V]
		args    args[K]
		want    V
		wantErr assert.ErrorAssertionFunc
	}
	tests := []testCase[string, string]{
		{
			name: "cache exist case",
			c: func() *cacheKV[string, string] {
				kv := New[string, string]()
				kv.cache.Set("key1", "val1")
				return kv
			}(),
			args:    args[string]{nil, "key1"},
			want:    "val1",
			wantErr: assert.NoError,
		}, {
			name:    "none source case",
			c:       New[string, string](),
			args:    args[string]{nil, "key1"},
			wantErr: assert.Error,
		}, {
			name: "source error case",
			c: New[string, string](WithSource(func() db.KV[string, string] {
				kv := mocks.MockKVStore[string, string]{
					GetFunc: func(ctx context.Context, k string) (string, error) { return "", errors.New("inner error") },
				}
				return kv
			}())),
			args:    args[string]{nil, "key1"},
			wantErr: assert.Error,
		}, {
			name: "source got case",
			c: New[string, string](WithSource(func() db.KV[string, string] {
				kv := mocks.MockKVStore[string, string]{
					GetFunc: func(ctx context.Context, k string) (string, error) { return "val1", nil },
				}
				return kv
			}())),
			want:    "val1",
			args:    args[string]{nil, "key1"},
			wantErr: assert.NoError,
		}, {
			name: "query from not found cache",
			c: func() *cacheKV[string, string] {
				kv := New(WithSource[string, string](mocks.MockKVStore[string, string]{
					GetFunc: func(ctx context.Context, k string) (string, error) { return "", fmt.Errorf("not found: %+v", k) },
				}))
				_, _ = kv.Get(context.Background(), "key1")
				return kv
			}(),
			args:    args[string]{nil, "key1"},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.c.Get(tt.args.ctx, tt.args.k)
			if !tt.wantErr(t, err, fmt.Sprintf("Get(%v, %v)", tt.args.ctx, tt.args.k)) {
				return
			}
			assert.Equalf(t, tt.want, got, "Get(%v, %v)", tt.args.ctx, tt.args.k)
		})
	}
}

func Test_cacheKV_cacheOptions(t *testing.T) {
	type testCase[K comparable, V any] struct {
		name string
		c    cacheKV[K, V]
		k    K
		want []cache.ItemOption
	}
	tests := []testCase[string, string]{
		{name: "none expire case", c: cacheKV[string, string]{}, k: "key1", want: nil},
		//  NOTE: got never equal to want expire-time cause expire-time equal to time.Now().Add(ttl)
		// {
		// 	name: "got never equal to want expire-time cause expire-time equal to time.Now().Add(ttl)",
		// 	c:    cacheKV[string, string]{ttl: time.Second},
		// 	want: []cache.ItemOption{cache.WithExpiration(time.Second)},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.c.cacheOptions(tt.k), "cacheOptions()")
		})
	}
}

func Test_cacheKV_Set(t *testing.T) {
	type args[K comparable, V any] struct {
		ctx context.Context
		k   K
		v   V
	}
	type testCase[K comparable, V any] struct {
		name    string
		c       *cacheKV[K, V]
		args    args[K, V]
		wantErr assert.ErrorAssertionFunc
	}
	tests := []testCase[string, string]{
		{
			name:    "all right case",
			c:       New[string, string](),
			args:    args[string, string]{nil, "key1", "val1"},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, tt.c.Set(tt.args.ctx, tt.args.k, tt.args.v), fmt.Sprintf("Set(%v, %v, %v)", tt.args.ctx, tt.args.k, tt.args.v))
		})
	}
}

func Test_cacheKV_Del(t *testing.T) {
	type args[K comparable] struct {
		ctx context.Context
		k   K
	}
	type testCase[K comparable, V any] struct {
		name    string
		c       cacheKV[K, V]
		args    args[K]
		wantErr assert.ErrorAssertionFunc
	}
	tests := []testCase[string, string]{
		{
			name:    "all right case",
			c:       *New[string, string](),
			args:    args[string]{nil, "key1"},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, tt.c.Del(tt.args.ctx, tt.args.k), fmt.Sprintf("Del(%v, %v)", tt.args.ctx, tt.args.k))
		})
	}
}
