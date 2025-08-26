package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

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
		Username    string
		ActivateURL string
	}{
		Username:    user.Username,
		ActivateURL: activateURL,
	}
	err = app.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !isProdEnv)
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
	//回写空的函数
	if err := app.jsonResponse(w, http.StatusCreated, userWIthToken); err != nil {
		app.internalServerError(w, r, err)
		return
	}

}
