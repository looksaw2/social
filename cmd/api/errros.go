package main

import (
	"log"
	"net/http"
)

// 服务器内部错误
func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal server error : %s path :%s error : %s", r.Method, r.URL.Path, err)
	writeJSONError(w, http.StatusInternalServerError, "The server encounter a problem")
}

// 请求错误返回
func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("bad request error : %s path :%s error : %s", r.Method, r.URL.Path, err)
	writeJSONError(w, http.StatusBadRequest, err.Error())
}

// Not Found Error
func (app *application) notFound(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("not Found error : %s path :%s error : %s", r.Method, r.URL.Path, err)
	writeJSONError(w, http.StatusNotFound, err.Error())
}
