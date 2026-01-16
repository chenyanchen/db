package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/chenyanchen/kv/cachekv"
)

func main() {
	ctx := context.Background()

	lru, err := cachekv.NewLRU[int64, Content](10, nil, 0)
	if err != nil {
		panic(err)
	}

	batchKV := cachekv.NewBatch[int64, Content](lru, &fakeContentKV{})

	contents, err := batchKV.Get(ctx, []int64{1, 3, 5})
	if err != nil {
		panic(err)
	}

	fmt.Println("contents:", contents)
	// output: contents: map[1:{1 Title: 1} 3:{3 Title: 3} 5:{5 Title: 5}]
}

type Content struct {
	ID    int64
	Title string
}

type fakeContentKV struct{}

func (s *fakeContentKV) Get(ctx context.Context, keys []int64) (map[int64]Content, error) {
	result := make(map[int64]Content, len(keys))
	for _, key := range keys {
		result[key] = Content{ID: key, Title: "Title: " + strconv.FormatInt(key, 10)}
	}
	return result, nil
}

func (s *fakeContentKV) Set(ctx context.Context, kvs map[int64]Content) error {
	return nil
}

func (s *fakeContentKV) Del(ctx context.Context, keys []int64) error {
	return nil
}
