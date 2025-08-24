package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

// 初始化
func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

// 写入json
func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

// 读取json
func readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	//最多读取的数据
	maxByte := 1_048_576
	//限制r.Body的长度
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxByte))
	decoder := json.NewDecoder(r.Body)
	//禁止未知的字段
	decoder.DisallowUnknownFields()
	return decoder.Decode(data)
}

// 写入错误json,发生错误时写入
func writeJSONError(w http.ResponseWriter, status int, message string) error {
	type envelope struct {
		Error string `json:"error"`
	}
	return writeJSON(w, status, &envelope{Error: message})
}

// 标准相应
func (app *application) jsonResponse(w http.ResponseWriter, status int, data any) error {
	type envelope struct {
		Data any `json:"data"`
	}
	return writeJSON(w, status, &envelope{Data: data})
}
