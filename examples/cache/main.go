// Example: Layered cache with database backend
//
// This example demonstrates how to compose a cache layer with a persistent store
// using layerkv. On cache miss, values are fetched from the store and cached.
package main

import (
	"context"
	"fmt"

	"github.com/chenyanchen/kv/cachekv"
	"github.com/chenyanchen/kv/examples/internal/mock"
	"github.com/chenyanchen/kv/layerkv"
)

func main() {
	ctx := context.Background()

	// 1. Create your database backend (implements kv.KV[int, *User])
	store := &mock.UserKV{}

	// 2. Create an LRU cache layer
	cache, err := cachekv.NewLRU[int, *mock.User](1000, nil, 0)
	if err != nil {
		panic(err)
	}

	// 3. Compose cache + store with layerkv
	//    - Cache-aside pattern: checks cache first, falls back to store
	//    - On cache miss, fetches from store and populates cache
	userKV, err := layerkv.New(cache, store)
	if err != nil {
		panic(err)
	}

	// First Get: cache miss, fetches from database
	user, err := userKV.Get(ctx, 1)
	if err != nil {
		panic(err)
	}
	fmt.Printf("1st get (cache miss): %+v\n", user)

	// Second Get: cache hit, returns from cache
	user, err = userKV.Get(ctx, 1)
	if err != nil {
		panic(err)
	}
	fmt.Printf("2nd get (cache hit): %+v\n", user)
}
