// Example: Full production composition
//
// This example demonstrates a production-ready setup combining multiple layers:
// 1. Database backend (your implementation)
// 2. Request deduplication (singleflightkv)
// 3. Telemetry for both cache and store
// 4. LRU cache with TTL
// 5. Layered composition with write-through
package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/chenyanchen/kv/cachekv"
	"github.com/chenyanchen/kv/examples/internal/mock"
	"github.com/chenyanchen/kv/examples/internal/telemetry"
	"github.com/chenyanchen/kv/layerkv"
	"github.com/chenyanchen/kv/singleflightkv"
)

func main() {
	ctx := context.Background()

	// 1. Database backend (your implementation)
	dbKV := &mock.UserKV{}

	// 2. Protect database with request deduplication
	//    Prevents thundering herd on cache miss
	dbKV2, err := singleflightkv.New[int, *mock.User](dbKV)
	if err != nil {
		panic(err)
	}

	// 3. Add telemetry to database
	dbWithMetrics := telemetry.Wrap(dbKV2, newRecorder("db"))

	// 4. LRU cache layer with TTL
	cache, err := cachekv.NewLRU[int, *mock.User](1000, nil, time.Minute*5)
	if err != nil {
		panic(err)
	}

	// 5. Add telemetry to cache
	cacheWithMetrics := telemetry.Wrap(cache, newRecorder("cache"))

	// 6. Compose: cache + store with write-through
	userKV, err := layerkv.New(cacheWithMetrics, dbWithMetrics, layerkv.WithWriteThrough())
	if err != nil {
		panic(err)
	}

	fmt.Println("=== Full Composition Example ===")
	fmt.Println()

	// Simulate concurrent requests for the same key
	// Only one will hit the database due to singleflight
	fmt.Println("First request (cache miss -> singleflight -> db):")
	user, _ := userKV.Get(ctx, 1)
	fmt.Printf("Result: %+v\n\n", user)

	fmt.Println("Second request (cache hit):")
	user, _ = userKV.Get(ctx, 1)
	fmt.Printf("Result: %+v\n\n", user)

	fmt.Println("Set with write-through (updates cache immediately):")
	_ = userKV.Set(ctx, 2, &mock.User{ID: 2, Name: "New User"})
	user, _ = userKV.Get(ctx, 2) // cache hit
	fmt.Printf("Result: %+v\n", user)
}

// Prometheus metrics
var histogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "example",
	Subsystem: "kv",
	Name:      "operation_duration_seconds",
}, []string{"layer", "operation", "success"})

func newRecorder(layer string) telemetry.RecordFunc {
	return func(operation string, success bool, duration time.Duration) {
		histogram.WithLabelValues(layer, operation, strconv.FormatBool(success)).
			Observe(duration.Seconds())
		fmt.Printf("  [%s] %s success=%v duration=%v\n", layer, operation, success, duration)
	}
}
