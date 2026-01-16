// Example: Basic cache usage
//
// This example demonstrates basic usage of the cachekv package
// with a simple LRU cache.
package main

import (
	"context"
	"fmt"

	"github.com/chenyanchen/kv/cachekv"
)

func main() {
	ctx := context.Background()

	// Create an LRU cache with capacity 10
	cache, err := cachekv.NewLRU[string, string](10, nil, 0)
	if err != nil {
		panic(err)
	}

	// Set a value
	if err := cache.Set(ctx, "greeting", "Hello, World!"); err != nil {
		panic(err)
	}

	// Get the value back
	value, err := cache.Get(ctx, "greeting")
	if err != nil {
		panic(err)
	}
	fmt.Println(value) // Output: Hello, World!

	// Delete the value
	if err := cache.Del(ctx, "greeting"); err != nil {
		panic(err)
	}

	// Get returns kv.ErrNotFound for missing keys
	_, err = cache.Get(ctx, "greeting")
	fmt.Println("After delete:", err) // Output: After delete: not found
}
