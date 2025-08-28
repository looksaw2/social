package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/looksaw/social/internal/mailer"
	"github.com/looksaw/social/internal/store"
)

// 请求时需要的
type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

// 返回时需要的
type UserWithToken struct {
	*store.User
	Token string `json:"token"`
}

// 用户注册的处理函数
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	//读取对应的payload
	var payload RegisterUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	//验证是否合规
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	//构成user
	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
	}
	//hash这个代码
	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(w, r, err)
		return
	}
	//存储user到数据库
	ctx := r.Context()
	//生成uuid
	plainToken := uuid.New().String()
	//sha加密
	hash := sha256.Sum256([]byte(plainToken))
	//转换成string
	hashToken := hex.EncodeToString(hash[:])
	err := app.store.Users.CreateAndInvite(ctx, user, hashToken, app.config.mail.exp)
	if err != nil {
		switch err {
		case store.ErrDuplicateUsername:
			app.badRequestResponse(w, r, err)
		case store.ErrDuplicateEmail:
			app.badRequestResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}
	userWIthToken := UserWithToken{
		User:  user,
		Token: plainToken,
	}
	//发送email
	activateURL := fmt.Sprintf("%s/confirm/%s", app.config.frontEndURL, plainToken)
	isProdEnv := app.config.env == "production"
	vars := struct {
		Username      string
		ActivationURL string
	}{
		Username:      user.Username,
		ActivationURL: activateURL,
	}
	status, err := app.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !isProdEnv)
	if err != nil {
		//SAGA
		app.logger.Errorw("error sending  welcome email", "error", err)
		//Rollback
		if err := app.store.Users.Delete(ctx, user.ID); err != nil {
			app.logger.Errorw("error deleting user", "error", err)
		}
		app.internalServerError(w, r, err)
		return
	}
	app.logger.Info("Email sent ", " status code ", status)
	//回写空的函数
	if err := app.jsonResponse(w, http.StatusCreated, userWIthToken); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}

// 发送的创建Token的请求
type CreateUserTokenPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

// 使用createToken生成token
func (app *application) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	// 读取验证发送的请求
	var payload CreateUserTokenPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	//通过邮件得到User
	user, err := app.store.Users.GetByEmail(r.Context(), payload.Email)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.unauthorizedBasicResponse(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
			return
		}
	}

	//创建一个Token
	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(app.config.auth.token.exp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": app.config.auth.token.iss,
		"aud": app.config.auth.token.iss,
	}
	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	//回写
	if err := app.jsonResponse(w, http.StatusCreated, token); err != nil {
		app.internalServerError(w, r, err)
	}
}

// 对于发送过来的Token进行验证
func (app *application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//得到头部
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorizedResponse(w, r, fmt.Errorf("authorization header is missing"))
			return
		}
		//解析头部
		parts := strings.Split(authHeader, " ") // authorization: Bearer <token>
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.unauthorizedResponse(w, r, fmt.Errorf("authorization header is malformed "))
			return
		}
		token := parts[1]
		//验证token
		jwtToken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			app.unauthorizedResponse(w, r, err)
			return
		}
		//得到claims
		claims, _ := jwtToken.Claims.(jwt.MapClaims)
		//得到userID
		userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)
		if err != nil {
			app.unauthorizedResponse(w, r, err)
			return
		}
		ctx := r.Context()
		user, err := app.store.Users.GetByID(ctx, userID)
		if err != nil {
			app.unauthorizedResponse(w, r, err)
			return
		}
		//写入上下文
		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
