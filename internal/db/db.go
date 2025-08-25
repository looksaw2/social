package db

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

// 初始化DB
func New(addr string, maxOpenConns int, maxIdleConns int, maxIdleTime string) (*sql.DB, error) {
	//设置postgresql链接
	db, err := sql.Open("postgres", addr)
	if err != nil {
		return nil, err
	}
	//设置cancel
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	//设置数据库的其他属性
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	//转换时间
	duration, err := time.ParseDuration(maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	//ping一下数据库
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil

}
