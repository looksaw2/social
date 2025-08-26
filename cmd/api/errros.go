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
