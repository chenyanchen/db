package mock

import (
	"context"
	"fmt"
)

// Content represents a content entity.
type Content struct {
	ID    int64
	Title string
}

// ContentBatchKV is a mock database implementation of kv.BatchKV[int64, Content].
// In production, this would be backed by a real database.
type ContentBatchKV struct {
	// db *sql.DB // uncomment to use real database
}

func (s *ContentBatchKV) Get(ctx context.Context, keys []int64) (map[int64]Content, error) {
	// Mock implementation - in production, batch query from database
	fmt.Printf("[mock] ContentBatchKV.Get(%v)\n", keys)
	result := make(map[int64]Content, len(keys))
	for _, key := range keys {
		result[key] = Content{ID: key, Title: fmt.Sprintf("Title %d", key)}
	}
	return result, nil
}

func (s *ContentBatchKV) Set(ctx context.Context, kvs map[int64]Content) error {
	// Mock implementation - in production, batch upsert to database
	fmt.Printf("[mock] ContentBatchKV.Set(%d items)\n", len(kvs))
	return nil
}

func (s *ContentBatchKV) Del(ctx context.Context, keys []int64) error {
	// Mock implementation - in production, batch delete from database
	fmt.Printf("[mock] ContentBatchKV.Del(%v)\n", keys)
	return nil
}
