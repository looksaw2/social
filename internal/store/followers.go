package store

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

// follower的模型
type Follower struct {
	UserID     int64  `json:"user_id"`
	FollowerID int64  `json:"follower_id"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

// follower的存储
type FollowerStorage struct {
	db *sql.DB
}

// Follow接口的实现
func (s *FollowerStorage) Follow(ctx context.Context, followerID int64, userID int64) error {
	query := `
		INSERT INTO followers (user_id , follower_id) VALUES ($1,$2)
	`
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	_, err := s.db.ExecContext(ctx, query, userID, followerID)
	//错误处理
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrConflict
		}
	}
	return nil
}

// Unfollow的实现
func (s *FollowerStorage) Unfollow(ctx context.Context, followerID int64, userID int64) error {
	query := `
		DELETE FROM followers
		WHERE user_id = $1 AND follower_id = $2
	`
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	_, err := s.db.ExecContext(ctx, query, userID, followerID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return ErrConflict
		}
	}
	return nil

}
