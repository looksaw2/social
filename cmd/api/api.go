package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"

	"github.com/looksaw/social/docs"
	"github.com/looksaw/social/internal/auth"
	"github.com/looksaw/social/internal/mailer"
	"github.com/looksaw/social/internal/store"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// application的struct，包含了必要的信息
type application struct {
	config        config             //配置属性
	store         *store.Storage     //存储属性
	logger        *zap.SugaredLogger //结构化的LOG
	mailer        mailer.Client      //发送mail的客户端
	authenticator auth.Authenticator //认证的类
}

// config的配置
type config struct {
	addr        string     //服务的端口
	db          dbConfig   //db的设置
	env         string     //是什么环境
	apiURL      string     //Swagger用的
	mail        mailConfig // mail的配置
	frontEndURL string     //前端的URL
	auth        authConfig //认证设计
}

type authConfig struct {
	basic basicConfig
	token tokenConfig
}

type tokenConfig struct {
	secret string
	exp    time.Duration
	iss    string
}

type basicConfig struct {
	username string
	pass     string
}

// mail的相关配置
type mailConfig struct {
	fromEmail string
	sendGrid  sendGridConfig
	mailTrip  mailTripConfig
	exp       time.Duration
}

// Send Grid的相关配置
type sendGridConfig struct {
	apiKey string
}

// mailTrip的配置
type mailTripConfig struct {
	apiKey string
}

// dbConfig
type dbConfig struct {
	addr         string //数据库地址
	maxOpenConns int    //数据库参数
	maxIdleConns int
	maxIdleTime  string
}

// 设置mux函数
func (app *application) mount() http.Handler {
	r := chi.NewRouter()
	//使用中间件
	//使用CORS
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	//使用路由
	r.Route("/v1", func(r chi.Router) {
		r.With(app.BasicAuthMiddleware()).Get("/health", app.healthCheck)
		//Swagger文档
		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

		//post路由
		r.Route("/posts", func(r chi.Router) {
			//进行Token的验证
			r.Use(app.AuthTokenMiddleware)
			//创建Post
			r.Post("/", app.createPostHandler)
			//得到对应ID的Post
			r.Route("/{postID}", func(r chi.Router) {
				//使用制作中间件
				r.Use(app.postContextMiddleware)
				//GET方法
				r.Get("/", app.getPostHandler)
				//Delete方法
				r.Delete("/", app.checkPostOwnership("admin", app.deletePostHandler))
				//PAtch方法
				r.Patch("/", app.checkPostOwnership("moderator", app.updatePostHandler))
			})
		})
		//User的路由
		r.Route("/users", func(r chi.Router) {
			r.Put("/activate/{token}", app.activateUserHandler)
			r.Route("/{userID}", func(r chi.Router) {
				//中间件
				r.Use(app.AuthTokenMiddleware)
				//得到用户信息
				r.Get("/", app.getUserHandler)
				//关注某人
				r.Put("/follow", app.followUserHandler)
				//取消关注某人
				r.Put("/unfollow", app.unFollowUserHandler)
			})
			//目前没有身份验证，姑且这样
			r.Group(func(r chi.Router) {
				r.Use(app.AuthTokenMiddleware)
				r.Get("/feed", app.getUserFeedHandler)
			})
		})
		//用户登陆注册
		r.Route("/authentication", func(r chi.Router) {
			//注册函数
			r.Post("/user", app.registerUserHandler)
			//得到token
			r.Post("/token", app.createTokenHandler)
		})
	})
	return r
}

// run函数，主要的执行函数
func (app *application) run(mux http.Handler) error {
	//添加swagger信息
	docs.SwaggerInfo.Version = VERSION

	docs.SwaggerInfo.Host = app.config.apiURL

	docs.SwaggerInfo.BasePath = "/v1"
	srv := &http.Server{
		Addr:         app.config.addr,  //设置INET4地址
		Handler:      mux,              //设置路由
		WriteTimeout: time.Second * 30, //设置写超时
		ReadTimeout:  time.Second * 10, //设置读超时
		IdleTimeout:  time.Minute,
	}
	//设置显示运行的信息
	app.logger.Infow("Start the server", "addr", app.config.addr, "env", app.config.env)
	return srv.ListenAndServe()
}
