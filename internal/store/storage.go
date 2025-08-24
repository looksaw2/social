package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound   = errors.New("resource not found")
	QueryDuration = time.Second * 5
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
	}
	//User接口
	Users interface {
		Create(context.Context, *User) error
	}
	//Comments接口
	Comment interface {
		GetPostByID(context.Context, int64) ([]Comment, error)
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
	}
}
