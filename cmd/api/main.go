package main

import (
	"log"
	"time"

	"github.com/looksaw/social/internal/auth"
	"github.com/looksaw/social/internal/db"
	"github.com/looksaw/social/internal/env"
	"github.com/looksaw/social/internal/mailer"
	"github.com/looksaw/social/internal/store"
	"go.uber.org/zap"
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
		//前端配置
		frontEndURL: env.GetString("FRONTEND_URL", "http://localhost:5173"),
		env:         env.GetString("ENV", "development"),
		//邮件配置
		mail: mailConfig{
			exp:       time.Hour * 24 * 3,
			fromEmail: env.GetString("FROM_EMAIL", "hello@demomailtrap.co"),
			sendGrid: sendGridConfig{
				apiKey: env.GetString("SENDGRID_API_KEY", ""),
			},
			//mailTrip设置
			mailTrip: mailTripConfig{
				apiKey: env.GetString("MAILTRIP_API_KEY", ""),
			},
		},
		//认证的基本设置
		auth: authConfig{
			basic: basicConfig{
				username: env.GetString("AUTH_BASIC_USERNAME", "admin"),
				pass:     env.GetString("AUTH_BASIC_PASS", "admin"),
			},
			token: tokenConfig{
				secret: env.GetString("AUTH_TOKEN_SECRET", "example"),
				exp:    time.Hour * 24 * 3,
				iss:    "gophersocial",
			},
		},
	}
	//初始化结构化logger
	logger := zap.Must(zap.NewProduction()).Sugar()
	defer logger.Sync()
	//初始化db
	db, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)
	if err != nil {
		logger.Fatalf("db setting is failed : %v", err)
	}
	defer db.Close()
	logger.Info("database connection pool established")
	//新建mail
	//Send grid mailer := mailer.NewSendgrid(cfg.mail.sendGrid.apiKey, cfg.mail.fromEmail)
	mailtrap, err := mailer.NewMailTrapClient(cfg.mail.mailTrip.apiKey, cfg.mail.fromEmail)
	//创建验证服务
	jwtAuthenticator := auth.NewJWTAuthenticator(cfg.auth.token.secret, cfg.auth.token.iss, cfg.auth.token.iss)

	if err != nil {
		logger.Fatal(err)
	}
	//初始化存储
	store := store.NewPostgreStorage(db)
	//初始化application
	app := &application{
		config:        cfg,
		store:         store,
		logger:        logger,
		mailer:        mailtrap,
		authenticator: jwtAuthenticator,
	}
	logger.Fatal(app.run(app.mount()))
}
