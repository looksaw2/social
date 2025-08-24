package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/looksaw/social/internal/store"
)

// 设置Postkey
type postKey string

var postCtx postKey = "post"

// 发送CreatePost的2请求结构体
type CreatePostPayload struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=1000"`
	Tags    []string `json:"tags"`
}

// 处理createPost的请求
func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreatePostPayload
	//读取结构体
	if err := readJSON(w, r, &payload); err != nil {
		app.internalServerError(w, r, err)
		return
	}
	//验证是否合规
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	post := &store.Post{
		//TODO
		UserID:  1,
		Title:   payload.Title,
		Content: payload.Content,
		Tags:    payload.Tags,
	}
	//得到对应的context
	ctx := r.Context()
	//写入Post
	if err := app.store.Posts.Create(ctx, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
	//返回写入Post
	if err := app.jsonResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// 得到对应ID的post
func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	//得到URL中的参数
	idParam := chi.URLParam(r, "postID")
	//解析
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	ctx := r.Context()
	//得到post
	post, err := app.store.Posts.GetByID(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFound(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}
	//得到Comments
	comments, err := app.store.Comment.GetPostByID(ctx, id)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	post.Comments = comments
	//写入post
	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}

// 删去操作
func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	//得到URL中的参数
	idParam := chi.URLParam(r, "postID")
	//解析
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	ctx := r.Context()

	if err := app.store.Posts.Delete(ctx, id); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFound(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

// Update的模型
type UpdatePayload struct {
	Title   *string `json:"title" validate:"omitempty,max=100"`
	Content *string `json:"content" validate:"omitempty,max=1000"`
}

// Update操作
func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	//从Context中得到post
	post := getPostFromCtx(r)

	//解析UpdatePayload
	var payload UpdatePayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	//验证payload是否合规
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	//将payload复制到post上面
	if payload.Content != nil {
		post.Content = *payload.Content
	}
	if payload.Title != nil {
		post.Title = *payload.Title
	}

	ctx := r.Context()
	//数据库执行Patch
	if err := app.store.Posts.Update(ctx, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// 实现关键功能的中间件，将得到的Post添加入上下文
func (app *application) postContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//得到URL中的参数
		idParam := chi.URLParam(r, "postID")
		//解析
		id, err := strconv.ParseInt(idParam, 10, 64)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
		ctx := r.Context()
		//得到post
		post, err := app.store.Posts.GetByID(ctx, id)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFound(w, r, err)
				return
			default:
				app.internalServerError(w, r, err)
				return
			}
		}
		//将Post存入ctx
		ctx = context.WithValue(ctx, postCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// 从r的上下文中间的到post
func getPostFromCtx(r *http.Request) *store.Post {
	post, _ := r.Context().Value(postCtx).(*store.Post)
	return post
}
