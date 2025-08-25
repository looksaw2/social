package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/looksaw/social/docs"
	"github.com/looksaw/social/internal/store"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// application的struct，包含了必要的信息
type application struct {
	config config         //配置属性
	store  *store.Storage //存储属性
}

// config的配置
type config struct {
	addr   string   //服务的端口
	db     dbConfig //db的设置
	env    string   //是什么环境
	apiURL string   //Swagger用的
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
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)
	//使用路由
	r.Route("/v1", func(r chi.Router) {
		r.Get("/health", app.healthCheck)
		//Swagger文档
		docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
		r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL(docsURL)))

		//post路由
		r.Route("/posts", func(r chi.Router) {
			//创建Post
			r.Post("/", app.createPostHandler)
			//得到对应ID的Post
			r.Route("/{postID}", func(r chi.Router) {
				//使用制作中间件
				r.Use(app.postContextMiddleware)
				//GET方法
				r.Get("/", app.getPostHandler)
				//Delete方法
				r.Delete("/", app.deletePostHandler)
				//PAtch方法
				r.Patch("/", app.updatePostHandler)
			})
		})
		//User的路由
		r.Route("/users", func(r chi.Router) {
			r.Route("/{userID}", func(r chi.Router) {
				//中间件
				r.Use(app.userContextMiddleware)
				//得到用户信息
				r.Get("/", app.getUserHandler)
				//关注某人
				r.Put("/follow", app.followUserHandler)
				//取消关注某人
				r.Put("/unfollow", app.unFollowUserHandler)
			})
			//目前没有身份验证，姑且这样
			r.Group(func(r chi.Router) {
				r.Get("/feed", app.getUserFeedHandler)
			})
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
	log.Printf("Start to run at port %s", app.config.addr)
	return srv.ListenAndServe()
}
