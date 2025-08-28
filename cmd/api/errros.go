package main

import (
	"net/http"
)

// 服务器内部错误
func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("internal server error :", "method", r.Method, "path", r.URL.Path, "err", err)
	writeJSONError(w, http.StatusInternalServerError, "The server encounter a problem")
}

// 请求错误返回
func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("bad request error :", "method", r.Method, "path", r.URL.Path, "err", err)
	writeJSONError(w, http.StatusBadRequest, err.Error())
}

// Not Found Error
func (app *application) notFound(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("not found error :", "method", r.Method, "path", r.URL.Path, "err", err)
	writeJSONError(w, http.StatusNotFound, err.Error())
}

// Conflict Response
func (app *application) conflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("conflict error :", "method", r.Method, "path", r.URL.Path, "err", err)
	writeJSONError(w, http.StatusConflict, err.Error())
}

// 认证失败
func (app *application) unauthorizedResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("unauthorized error :", "method", r.Method, "path", r.URL.Path, "err", err)
	writeJSONError(w, http.StatusNonAuthoritativeInfo, err.Error())
}

// Basic 认证失败
func (app *application) unauthorizedBasicResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("unauthorized basic error :", "method", r.Method, "path", r.URL.Path, "err", err)
	//查看MDN的文档去
	w.Header().Set("WWW-Authenticate", `Basic realm="restricted" ,charset="UTF-8"`)
	writeJSONError(w, http.StatusNonAuthoritativeInfo, err.Error())
}

// forbidden
func (app *application) forbiddenResponse(w http.ResponseWriter, r *http.Request) {
	app.logger.Warnf("forbidden error :", "method", r.Method, "path", r.URL.Path, "err", err)
	writeJSONError(w, http.StatusForbidden, "forbidden")
}
