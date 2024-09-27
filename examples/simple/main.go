package main

import (
	"context"
	"fmt"

	"github.com/chenyanchen/db/cachekv"
)

func main() {
	kv, err := cachekv.NewLRU[string, string](10, nil, 0)
	if err != nil {
		panic(err)
	}
	ctx := context.TODO()
	if err := kv.Set(ctx, "foo", "bar"); err != nil {
		panic(err)
	}
	fmt.Println(kv.Get(ctx, "foo"))
	// Output: bar <nil>
}
