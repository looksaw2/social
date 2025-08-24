package store

import (
	"context"
	"database/sql"
)

// User模型
type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"-"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// User存储
type UserStore struct {
	db *sql.DB
}

// Create方法实现
func (s *UserStore) Create(ctx context.Context, user *User) error {
	query :=
		`
		INSERT INTO users (username , email , password)
		VALUES ($1,$2,$3) RETURNING id,created_at,updated_at
	`
	err := s.db.QueryRowContext(
		ctx,
		query,
		user.Username,
		user.Email,
		user.Password,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}
