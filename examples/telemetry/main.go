// Example: Adding telemetry to KV operations
//
// This example demonstrates how to wrap kv.KV implementations with
// telemetry to record operation metrics using Prometheus.
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
)

func main() {
	ctx := context.Background()

	// 1. Create database backend
	store := &mock.UserKV{}

	// 2. Create LRU cache
	cache, err := cachekv.NewLRU[int, *mock.User](1000, nil, 0)
	if err != nil {
		panic(err)
	}

	// 3. Wrap both with telemetry to track metrics separately
	//    This allows calculating cache hit ratio:
	//    rate(cache_get_success) / (rate(cache_get_success) + rate(store_get_success))
	cacheWithMetrics := telemetry.Wrap(cache, newRecorder("cache"))
	storeWithMetrics := telemetry.Wrap(store, newRecorder("store"))

	// 4. Compose with layerkv
	userKV, err := layerkv.New(cacheWithMetrics, storeWithMetrics)
	if err != nil {
		panic(err)
	}

	// Operations will now record metrics
	user, _ := userKV.Get(ctx, 1) // cache miss, store hit
	fmt.Printf("1st get: %+v\n", user)

	user, _ = userKV.Get(ctx, 1) // cache hit
	fmt.Printf("2nd get: %+v\n", user)

	// In production, expose metrics via HTTP:
	// http.Handle("/metrics", promhttp.Handler())
	// http.ListenAndServe(":8080", nil)
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
		fmt.Printf("[%s] %s success=%v duration=%v\n", layer, operation, success, duration)
	}
}
