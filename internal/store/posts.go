package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

// post模型
type Post struct {
	ID        int64    `json:"id"`
	Content   string   `json:"content"`
	Title     string   `json:"title"`
	UserID    int64    `json:"user_id"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`

	//连接的外表
	Comments []Comment `json:"comment"`
	//乐观锁
	Version int64 `json:"version"`
}

// post存储
type PostStore struct {
	db *sql.DB
}

// 实现Create接口
func (s *PostStore) Create(ctx context.Context, post *Post) error {
	//SQL语句
	query := `
		INSERT INTO posts(content , title , user_id , tags)
		VALUES($1,$2,$3,$4) RETURNING id ,created_at ,updated_at
	`
	//超时控制
	ctx, cancel := context.WithTimeout(ctx, QueryDuration)
	defer cancel()
	//开始查询
	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Content,
		post.Title,
		post.UserID,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}

// 实现getById接口 posts
func (s *PostStore) GetByID(ctx context.Context, id int64) (*Post, error) {
	query :=
		`
		SELECT id , user_id , title , created_at , updated_at , tags ,version
		FROM posts WHERE id = $1
	`
	//超时控制
	ctx, cancel := context.WithTimeout(ctx, QueryDuration)
	defer cancel()
	//执行
	var post Post
	err := s.db.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.CreatedAt,
		&post.UpdatedAt,
		pq.Array(&post.Tags),
		&post.Version,
	)
	//错误处理
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, err
		default:
			return nil, err
		}
	}
	return &post, nil
}

// Delete方法
func (s *PostStore) Delete(ctx context.Context, postID int64) error {
	//请求
	query :=
		`
		DELETE FROM posts WHERE id = $1
	`
	//超时控制
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	//执行删除操作
	res, err := s.db.ExecContext(ctx, query, postID)
	if err != nil {
		return err
	}
	//查看删除了多少行
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	//删除0行，没有找到
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// Patch方法
func (s *PostStore) Update(ctx context.Context, post *Post) error {
	query := `
		UPDATE posts
		SET title = $1 , content = $2 , version = version + 1
		WHERE id = $3 AND version = $4
		RETURNING version
	`
	//超时控制
	ctx, cancel := context.WithTimeout(ctx, QueryDuration)
	defer cancel()
	//执行Patch操作
	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Content,
		post.ID,
		post.Version).Scan(
		&post.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return ErrNotFound
		default:
			return err
		}
	}
	return nil
}
