package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/looksaw/social/internal/store"
)

// 基础的Basic 认证
func (app *application) BasicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			//读取文件头
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.unauthorizedBasicResponse(w, r, fmt.Errorf("authorization headers is misssing"))
				return
			}
			//解析文件头部
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Basic" {
				app.unauthorizedBasicResponse(w, r, fmt.Errorf("authorization header is malformed"))
				return
			}
			//对得到的Basic认证信息解码(base64)
			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {
				app.unauthorizedBasicResponse(w, r, err)
				return
			}
			//得到配置中的username ，pass
			username := app.config.auth.basic.username
			pass := app.config.auth.basic.pass

			creds := strings.SplitN(string(decoded), ":", 2)
			if len(creds) != 2 || creds[0] != username || creds[1] != pass {
				app.unauthorizedBasicResponse(w, r, fmt.Errorf("invalid credentials"))
				return
			}

			//结束中间件
			next.ServeHTTP(w, r)
		})
	}
}

// 检查权限
func (app *application) checkPostOwnership(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromContext(r)
		post := getPostFromCtx(r)
		//是不是自己的帖子
		if post.UserID == user.ID {
			next.ServeHTTP(w, r)
			return
		}
		//检查权限
		allowed, err := app.checkRolePrecedence(r.Context(), user, requiredRole)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
		if !allowed {
			app.followUserHandler(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// 检查权限
func (app *application) checkRolePrecedence(ctx context.Context, user *store.User, roleName string) (bool, error) {
	role, err := app.store.Roles.GetByName(ctx, roleName)
	if err != nil {
		return false, err
	}
	return user.Role.Level > role.Level, nil
}
