package main

import (
	"log"

	"github.com/looksaw/social/internal/db"
	"github.com/looksaw/social/internal/env"
	"github.com/looksaw/social/internal/store"
)

//将随机生成的数据写入数据库

func main() {
	//得到数据库的Uri
	addr := env.GetString("DB_ADDR", "postgres://postgres:postgres@localhost:5432/social?sslmode=disable")
	//新建数据库连接
	conn, err := db.New(addr, 30, 30, "15m")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	store := store.NewPostgreStorage(conn)
	db.Seed(store)
}
