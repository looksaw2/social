package main

import (
	"log"

	"github.com/looksaw/social/internal/db"
	"github.com/looksaw/social/internal/env"
	"github.com/looksaw/social/internal/store"
)

// 版本号
const VERSION = "0.0.1"

//	@title			GopherSocial API
//	@description	This is a API for Gophersocial.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/v2
//
// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description
func main() {
	//读取环境变量
	if err := env.Init(); err != nil {
		log.Fatalf("Read the .env file failed %v", err)
		return
	}
	//初始化config
	cfg := config{
		addr:   env.GetString("ADDR", ":8080"),
		apiURL: env.GetString("EXTERNAL_URL", "localhost:8080"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://postgres:postgres@localhost:5432/social?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env: env.GetString("ENV", "development"),
	}
	//初始化db
	db, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)
	if err != nil {
		log.Fatalf("db setting is failed : %v", err)
	}
	defer db.Close()
	//初始化存储
	store := store.NewPostgreStorage(db)
	//初始化application
	app := &application{
		config: cfg,
		store:  store,
	}
	log.Fatal(app.run(app.mount()))
}
