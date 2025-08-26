package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("resource not found")
	QueryDuration        = time.Second * 5
	ErrConflict          = errors.New("resource already exists")
	ErrDuplicateEmail    = errors.New("duplicate email")
	ErrDuplicateUsername = errors.New("duplicate username")
)

type Storage struct {
	//Posts接口
	Posts interface {
		//GET请求
		GetByID(context.Context, int64) (*Post, error)
		//POST请求
		Create(context.Context, *Post) error
		//PATCH请求
		Update(context.Context, *Post) error
		//DELETE请求
		Delete(context.Context, int64) error
		//
		GetUserFeed(context.Context, int64, PaginationFeedQuery) ([]PostWithMetadata, error)
	}
	//User接口
	Users interface {
		Create(context.Context, *sql.Tx, *User) error
		GetByID(context.Context, int64) (*User, error)
		CreateAndInvite(context.Context, *User, string, time.Duration) error
		Activate(context.Context, string) error
		Delete(context.Context, int64) error
	}
	//Comments接口
	Comment interface {
		//通过ID获取评论
		GetPostByID(context.Context, int64) ([]Comment, error)
		//创建评论
		Create(context.Context, *Comment) error
	}
	Followers interface {
		//关注某人
		Follow(context.Context, int64, int64) error
		Unfollow(context.Context, int64, int64) error
	}
}

// 初始化PG存储
func NewPostgreStorage(db *sql.DB) *Storage {
	return &Storage{
		Posts: &PostStore{
			db: db,
		},
		Users: &UserStore{
			db: db,
		},
		Comment: &CommentsStore{
			db: db,
		},
		Followers: &FollowerStorage{
			db: db,
		},
	}
}

// 数据库事务
func withTx(db *sql.DB, ctx context.Context, fn func(*sql.Tx) error) error {
	//创建事务
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	//执行事务
	if err := fn(tx); err != nil {
		//失败回滚
		_ = tx.Rollback()
		return err
	}
	//提交
	return tx.Commit()
}
