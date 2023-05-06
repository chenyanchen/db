package db

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"git.in.zhihu.com/chenyanchen/db/mocks"
)

func Test_sfKV_Get(t *testing.T) {
	type args[K comparable] struct {
		ctx context.Context
		k   K
	}
	type testCase[K comparable, V any] struct {
		name    string
		s       *sfKV[K, V]
		args    args[K]
		want    V
		wantErr assert.ErrorAssertionFunc
	}
	tests := []testCase[string, string]{
		{
			name: "error case",
			s: func() *sfKV[string, string] {
				kv, err := NewSingleFlightKV[string, string](mocks.MockKVStore[string, string]{
					GetFunc: func(ctx context.Context, k string) (string, error) { return "", errors.New("not found") },
				})
				assert.NoError(t, err)
				return kv
			}(),
			args:    args[string]{nil, "not exist key"},
			wantErr: assert.Error,
		}, {
			name: "right case",
			s: func() *sfKV[string, string] {
				kv, err := NewSingleFlightKV[string, string](mocks.MockKVStore[string, string]{
					GetFunc: func(ctx context.Context, k string) (string, error) { return "val1", nil },
				})
				assert.NoError(t, err)
				return kv
			}(),
			args:    args[string]{nil, "key1"},
			want:    "val1",
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		got, err := tt.s.Get(tt.args.ctx, tt.args.k)
		if !tt.wantErr(t, err, fmt.Sprintf("Get(%v, %v)", tt.args.ctx, tt.args.k)) {
			return
		}
		assert.Equalf(t, tt.want, got, "Get(%v, %v)", tt.args.ctx, tt.args.k)
	}
}

func Test_sfKV_Set(t *testing.T) {
	type args[K comparable, V any] struct {
		ctx context.Context
		k   K
		v   V
	}
	type testCase[K comparable, V any] struct {
		name    string
		s       *sfKV[K, V]
		args    args[K, V]
		wantErr assert.ErrorAssertionFunc
	}
	tests := []testCase[string, string]{
		{
			name: "error case ",
			s: func() *sfKV[string, string] {
				kv, err := NewSingleFlightKV[string, string](mocks.MockKVStore[string, string]{
					SetFunc: func(ctx context.Context, k, v string) error { return errors.New("inner error") },
				})
				assert.NoError(t, err)
				return kv
			}(),
			args:    args[string, string]{nil, "key", "not exist key"},
			wantErr: assert.Error,
		}, {
			name: "right case",
			s: func() *sfKV[string, string] {
				kv, err := NewSingleFlightKV[string, string](mocks.MockKVStore[string, string]{
					SetFunc: func(ctx context.Context, k, v string) error { return nil },
				})
				assert.NoError(t, err)
				return kv
			}(),
			args:    args[string, string]{nil, "key", "not exist key"},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, tt.s.Set(tt.args.ctx, tt.args.k, tt.args.v), fmt.Sprintf("Set(%v, %v, %v)", tt.args.ctx, tt.args.k, tt.args.v))
		})
	}
}

func Test_sfKV_Del(t *testing.T) {
	type args[K comparable] struct {
		ctx context.Context
		k   K
	}
	type testCase[K comparable, V any] struct {
		name    string
		s       *sfKV[K, V]
		args    args[K]
		wantErr assert.ErrorAssertionFunc
	}
	tests := []testCase[string, string]{
		{
			name: "error case",
			s: func() *sfKV[string, string] {
				kv, err := NewSingleFlightKV[string, string](mocks.MockKVStore[string, string]{
					DelFunc: func(ctx context.Context, k string) error { return errors.New("inner error") },
				})
				assert.NoError(t, err)
				return kv
			}(),
			args:    args[string]{nil, "key"},
			wantErr: assert.Error,
		}, {
			name: "right case",
			s: func() *sfKV[string, string] {
				kv, err := NewSingleFlightKV[string, string](mocks.MockKVStore[string, string]{
					DelFunc: func(ctx context.Context, k string) error { return nil },
				})
				assert.NoError(t, err)
				return kv
			}(),
			args:    args[string]{nil, "key"},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, tt.s.Del(tt.args.ctx, tt.args.k), fmt.Sprintf("Del(%v, %v)", tt.args.ctx, tt.args.k))
		})
	}
}
