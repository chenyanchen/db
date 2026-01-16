// Example: Batch operations with cache
//
// This example demonstrates how to use cachekv.NewBatch for batch operations.
// It combines a cache layer with a batch-capable store to efficiently
// handle multiple keys at once.
package main

import (
	"context"
	"fmt"

	"github.com/chenyanchen/kv/cachekv"
	"github.com/chenyanchen/kv/examples/internal/mock"
)

func main() {
	ctx := context.Background()

	// 1. Create an LRU cache for Content
	cache, err := cachekv.NewLRU[int64, mock.Content](100, nil, 0)
	if err != nil {
		panic(err)
	}

	// 2. Create your batch-capable database backend (implements kv.BatchKV)
	store := &mock.ContentBatchKV{}

	// 3. Combine cache with batch store
	//    - On batch Get, checks cache first for each key
	//    - Fetches missing keys from store in a single batch
	//    - Populates cache with fetched values
	batchKV := cachekv.NewBatch(cache, store)

	// Batch Get: fetches multiple keys at once
	contents, err := batchKV.Get(ctx, []int64{1, 3, 5})
	if err != nil {
		panic(err)
	}
	fmt.Println("batch get (all cache miss):", contents)

	// Second batch Get: some keys now in cache
	contents, err = batchKV.Get(ctx, []int64{1, 2, 3})
	if err != nil {
		panic(err)
	}
	fmt.Println("batch get (partial cache hit):", contents)
}
