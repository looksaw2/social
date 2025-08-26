package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/looksaw/social/internal/store"
)

type userKey string

var userCtx userKey = "user"

type FollwerUser struct {
	UserID int64 `json:"user_id"`
}

// 处理/users/{userID}的GET请求

// GetUser godoc
//
//	@Summary		Fetches a user profile
//	@Description	get string by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	store.User
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Router			/users/{id} [get]
func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	//利用中间件的信息
	user := getUserFromContext(r)
	//返回结果
	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}

// 关注某人的API
func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromContext(r)
	//临时的措施
	var payload FollwerUser
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	//写入followers表
	ctx := r.Context()
	if err := app.store.Followers.Follow(ctx, followerUser.ID, payload.UserID); err != nil {
		switch err {
		case store.ErrConflict:
			app.conflictResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}
	//回写
	if err := app.jsonResponse(w, http.StatusOK, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// 解除对某人的关注
func (app *application) unFollowUserHandler(w http.ResponseWriter, r *http.Request) {
	unfollowerUser := getUserFromContext(r)
	if err := app.jsonResponse(w, http.StatusOK, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
	//临时的措施
	var payload FollwerUser
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	//写入followers表
	ctx := r.Context()
	if err := app.store.Followers.Unfollow(ctx, unfollowerUser.ID, payload.UserID); err != nil {
		switch err {
		case store.ErrConflict:
			app.conflictResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}
	//回写

	if err := app.jsonResponse(w, http.StatusOK, nil); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// 得到UserID的中间件
func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//得到UserIDParam
		userIDParam := chi.URLParam(r, "userID")
		if userIDParam == "" {
			app.badRequestResponse(w, r, errors.New("userID is required"))
			return
		}
		//解析userID
		userID, err := strconv.ParseInt(userIDParam, 10, 64)
		if err != nil {
			app.badRequestResponse(w, r, err)
			return
		}
		//得到对应的user
		ctx := r.Context()
		user, err := app.store.Users.GetByID(ctx, userID)
		if err != nil {
			switch err {
			case store.ErrNotFound:
				app.notFound(w, r, err)
				return
			default:
				app.internalServerError(w, r, err)
				return
			}
		}
		//将得到user添加进入context
		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// 从context中得到user
func getUserFromContext(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtx).(*store.User)
	return user
}

// 激活用户
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	//从URL中得到token
	token := chi.URLParam(r, "token")
	err := app.store.Users.Activate(r.Context(), token)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFound(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}
	//回写
	if err := app.jsonResponse(w, http.StatusNoContent, ""); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}
