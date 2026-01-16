package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	kv "github.com/chenyanchen/kv"
	"github.com/chenyanchen/kv/cachekv"
	"github.com/chenyanchen/kv/layerkv"
)

func main() {
	userDatabaseKV := &databaseKV{}
	userLRUKV, err := cachekv.NewLRU[int, *User](1<<10, nil, 0)
	if err != nil {
		panic(err)
	}

	// For example cache hit ratio:
	// 	rate(example_kv_user_kv_operation_duration_seconds_count{name="user_lru_kv", operation="Get", success="true"}[$__rate_interval])) /
	// 	(rate(example_kv_user_kv_operation_duration_seconds_count{name="user_lru_kv", operation="Get", success="true"}[$__rate_interval]) + rate(example_kv_user_kv_operation_duration_seconds_count{name="user_database_kv", operation="Get", success="true"}[$__rate_interval]))
	userKV, err := layerkv.New(
		NewTelemetry(userLRUKV, NewRecorder("user_lru_kv")),
		NewTelemetry(userDatabaseKV, NewRecorder("user_database_kv")),
	)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// 1st get user from database
	user, err := userKV.Get(ctx, 1)
	if err != nil {
		panic(err)
	}

	// 2nd get user from cache
	user, err = userKV.Get(ctx, 1)
	if err != nil {
		panic(err)
	}

	fmt.Printf("user: %+v\n", user)
}

type User struct {
	ID   int
	Name string
}

// Database implementation

type databaseKV struct {
	// uncomment the following line to use the database
	// db *sql.DB
}

func (s *databaseKV) Get(ctx context.Context, id int) (*User, error) {
	return &User{
		ID:   id,
		Name: "Mock Name",
	}, nil
}

func (s *databaseKV) Set(ctx context.Context, id int, user *User) error {
	return nil
}

func (s *databaseKV) Del(ctx context.Context, id int) error {
	return nil
}

// Telemetry implementation

type recordFunc func(operation string, success bool, duration time.Duration)

type telemetry[K comparable, V any] struct {
	next   kv.KV[K, V]
	record recordFunc
}

func NewTelemetry[K comparable, V any](next kv.KV[K, V], record recordFunc) telemetry[K, V] {
	return telemetry[K, V]{next: next, record: record}
}

func (t telemetry[K, V]) Get(ctx context.Context, k K) (V, error) {
	now := time.Now()
	v, err := t.next.Get(ctx, k)
	t.record("Get", err == nil, time.Since(now))
	return v, err
}

func (t telemetry[K, V]) Set(ctx context.Context, k K, v V) error {
	now := time.Now()
	err := t.next.Set(ctx, k, v)
	t.record("Set", err == nil, time.Since(now))
	return err
}

func (t telemetry[K, V]) Del(ctx context.Context, k K) error {
	now := time.Now()
	err := t.next.Del(ctx, k)
	t.record("Del", err == nil, time.Since(now))
	return err
}

var histogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "example",
	Subsystem: "kv",
	Name:      "user_kv_operation_duration_seconds",
}, []string{"name", "operation", "success"})

func NewRecorder(name string) recordFunc {
	return func(operation string, success bool, duration time.Duration) {
		histogram.WithLabelValues(name, operation, strconv.FormatBool(success)).Observe(duration.Seconds())
	}
}
