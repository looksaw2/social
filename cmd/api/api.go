package main

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/looksaw/social/internal/store"
	"log"
	"net/http"
	"time"
)

// application的struct，包含了必要的信息
type application struct {
	config config         //配置属性
	store  *store.Storage //存储属性
}

// config的配置
type config struct {
	addr string   //服务的端口
	db   dbConfig //db的设置
	env  string   //是什么环境
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
	})
	return r
}

// run函数，主要的执行函数
func (app *application) run(mux http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,  //设置INET4地址
		Handler:      mux,              //设置路由
		WriteTimeout: time.Second * 30, //设置写超时
		ReadTimeout:  time.Second * 10, //设置读超时
		IdleTimeout:  time.Minute,
	}
	//设置显示运行的信息
	log.Printf("Start to tun at port %s", app.config.addr)
	return srv.ListenAndServe()
}
