// Example: Sharded cache for high concurrency
//
// This example demonstrates the sharded cache implementation which
// provides better performance under high concurrent access by reducing
// lock contention through sharding.
package main

import (
	"context"
	"fmt"

	"github.com/chenyanchen/kv/cachekv"
)

func main() {
	ctx := context.Background()

	// Create a sharded cache with 32 shards
	// More shards = less lock contention = better concurrent performance
	const numShards = 32

	// Works with any comparable key type

	// Example 1: String keys
	stringCache := cachekv.NewSharded[string, string](numShards)
	_ = stringCache.Set(ctx, "foo", "bar")
	v1, _ := stringCache.Get(ctx, "foo")
	fmt.Println("string key:", v1)

	// Example 2: Integer keys
	intCache := cachekv.NewSharded[int, string](numShards)
	_ = intCache.Set(ctx, 42, "answer")
	v2, _ := intCache.Get(ctx, 42)
	fmt.Println("int key:", v2)

	// Example 3: Struct keys (must be comparable)
	type Point struct{ X, Y int }
	pointCache := cachekv.NewSharded[Point, string](numShards)
	_ = pointCache.Set(ctx, Point{1, 2}, "origin")
	v3, _ := pointCache.Get(ctx, Point{1, 2})
	fmt.Println("struct key:", v3)
}
