package main

import (
	"context"
	"fmt"
	"time"

	"git.in.zhihu.com/chenyanchen/db"
)

func main() {
	kv := db.NewCacheKV[string, string](db.WithSmoothExpires[string, string](time.Hour))
	ctx := context.TODO()
	if err := kv.Set(ctx, "foo", "bar"); err != nil {
		panic(err)
	}
	fmt.Println(kv.Get(ctx, "foo"))
	// Output: bar <nil>
}
