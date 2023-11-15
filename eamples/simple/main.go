package main

import (
	"context"
	"fmt"
	"time"

	"github.com/chenyanchen/db/cachekv"
)

func main() {
	kv := cachekv.New[string, string](cachekv.WithSmoothExpires[string, string](time.Hour))
	ctx := context.TODO()
	if err := kv.Set(ctx, "foo", "bar"); err != nil {
		panic(err)
	}
	fmt.Println(kv.Get(ctx, "foo"))
	// Output: bar <nil>
}
