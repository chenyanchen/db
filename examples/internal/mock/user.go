package mock

import (
	"context"
	"fmt"
)

// User represents a user entity.
type User struct {
	ID   int
	Name string
}

// UserKV is a mock database implementation of kv.KV[int, *User].
// In production, this would be backed by a real database.
type UserKV struct {
	// db *sql.DB // uncomment to use real database
}

func (s *UserKV) Get(ctx context.Context, id int) (*User, error) {
	// Mock implementation - in production, query from database:
	// var user User
	// err := s.db.QueryRowContext(ctx, "SELECT id, name FROM users WHERE id = ?", id).
	//     Scan(&user.ID, &user.Name)
	// if errors.Is(err, sql.ErrNoRows) {
	//     return nil, kv.ErrNotFound
	// }
	// return &user, err
	fmt.Printf("[mock] UserKV.Get(%d)\n", id)
	return &User{
		ID:   id,
		Name: fmt.Sprintf("User %d", id),
	}, nil
}

func (s *UserKV) Set(ctx context.Context, id int, user *User) error {
	// Mock implementation - in production, upsert to database
	fmt.Printf("[mock] UserKV.Set(%d, %+v)\n", id, user)
	return nil
}

func (s *UserKV) Del(ctx context.Context, id int) error {
	// Mock implementation - in production, delete from database
	fmt.Printf("[mock] UserKV.Del(%d)\n", id)
	return nil
}
