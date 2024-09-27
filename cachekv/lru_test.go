package cachekv

import (
	"context"
	"fmt"
	"testing"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/hashicorp/golang-lru/v2/simplelru"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_lruKV_Get(t *testing.T) {
	type args[K comparable] struct {
		ctx context.Context
		k   K
	}
	type testCase[K comparable, V any] struct {
		name    string
		c       lruKV[K, V]
		args    args[K]
		want    V
		wantErr assert.ErrorAssertionFunc
	}
	tests := []testCase[string, string]{
		{
			name: "exist",
			c: lruKV[string, string]{
				cache: func() simplelru.LRUCache[string, string] {
					cache, err := lru.New[string, string](2)
					require.NoError(t, err)
					cache.Add("key1", "value1")
					return cache
				}(),
			},
			args: args[string]{
				ctx: context.Background(),
				k:   "key1",
			},
			want:    "value1",
			wantErr: assert.NoError,
		}, {
			name: "not exist",
			c: lruKV[string, string]{
				cache: expirable.NewLRU[string, string](2, nil, 0),
			},
			args: args[string]{
				ctx: context.Background(),
				k:   "key1",
			},
			want:    "",
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
