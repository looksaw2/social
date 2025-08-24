package main

import (
	"net/http"
)

// healthCheck的路由
func (app *application) healthCheck(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status":  "ok",
		"env":     app.config.env,
		"version": VERSION,
	}
	//写入json的错误
	if err := app.jsonResponse(w, http.StatusOK, data); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
	}
}
