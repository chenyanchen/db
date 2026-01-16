package main

import (
	"context"
	"fmt"

	"github.com/chenyanchen/kv/cachekv"
)

type Point struct {
	X, Y int
}

func main() {
	ctx := context.Background()

	const numShards = 32

	// NewSharded internally uses maphash.Comparable(seed, key) to pick a shard.
	// Below are equivalent "hash examples" for int/string/struct keys.

	// string
	{
		key := "foo"
		kv := cachekv.NewSharded[string, string](numShards)
		_ = kv.Set(ctx, key, "bar")
		v, _ := kv.Get(ctx, key)
		fmt.Println("string value:", v)
	}

	// int
	{
		key := 42
		kv := cachekv.NewSharded[int, string](numShards)
		_ = kv.Set(ctx, key, "answer")
		v, _ := kv.Get(ctx, key)
		fmt.Println("int value:", v)
	}

	// struct (Point)
	{
		key := Point{X: 1, Y: 2}
		kv := cachekv.NewSharded[Point, string](numShards)
		_ = kv.Set(ctx, key, "pt")
		v, _ := kv.Get(ctx, Point{X: 1, Y: 2})
		fmt.Println("point value:", v)
	}
}
