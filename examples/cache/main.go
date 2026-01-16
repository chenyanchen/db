package main

import (
	"context"
	"fmt"

	"github.com/chenyanchen/kv/cachekv"
	"github.com/chenyanchen/kv/layerkv"
)

func main() {
	userDatabaseKV := &databaseKV{}
	userLRUKV, err := cachekv.NewLRU[int, *User](1<<10, nil, 0)
	if err != nil {
		panic(err)
	}

	userKV, err := layerkv.New(userLRUKV, userDatabaseKV)
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
