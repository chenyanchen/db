package main

import (
	"context"
	"fmt"

	"github.com/chenyanchen/db"
)

func main() {
	// You can define K as all comparable types, like this:
	// array: [3]int64
	// string (SQL): "id in (1, 2, 3)"
	// string (JSON): `{"ids": [1, 2, 3]}`
	var mapKV db.KV[[3]int64, []Content] = NewMapKV()

	if err := mapKV.Set(nil, [3]int64{}, []Content{{ID: 1, Title: "A"}, {ID: 2, Title: "B"}}); err != nil {
		panic(err)
	}

	contents, err := mapKV.Get(nil, [3]int64{1, 3})
	if err != nil {
		panic(err)
	}

	fmt.Println("contents:", contents)
	// output: contents: [{1 A}]
}

type Content struct {
	ID    int64
	Title string
}

type mapKV struct {
	m map[int64]Content
}

func NewMapKV() *mapKV {
	return &mapKV{m: make(map[int64]Content)}
}

func (s *mapKV) Get(ctx context.Context, ids [3]int64) ([]Content, error) {
	var contents []Content
	for _, id := range ids {
		content, ok := s.m[id]
		if !ok {
			continue
		}
		contents = append(contents, content)
	}
	return contents, nil
}

func (s *mapKV) Set(ctx context.Context, ids [3]int64, contents []Content) error {
	for _, content := range contents {
		s.m[content.ID] = content
	}
	return nil
}

func (s *mapKV) Del(ctx context.Context, ids [3]int64) error {
	for _, id := range ids {
		delete(s.m, id)
	}
	return nil
}
