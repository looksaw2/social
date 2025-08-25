package main

import (
	"net/http"

	"github.com/looksaw/social/internal/store"
)

func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	//处理分页
	//默认的参数
	fq := store.PaginationFeedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
	}
	//处理有可能传过来的参数
	fq, err := fq.Parse(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	//验证参数是否靠谱
	if err := Validate.Struct(fq); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	//得到ctx
	ctx := r.Context()
	feed, err := app.store.Posts.GetUserFeed(ctx, int64(1), fq)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	//回写
	if err := app.jsonResponse(w, http.StatusOK, feed); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
