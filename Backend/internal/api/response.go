package api

import (
	"encoding/json"
	"net/http"
	"time"
)

type Response struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Data       any    `json:"data"`
	ServerTime string `json:"serverTime"`
}

func Success(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, Response{
		Code:       CodeOK,
		Message:    "ok",
		Data:       data,
		ServerTime: time.Now().Format(time.RFC3339),
	})
}

func Created(w http.ResponseWriter, data any) {
	Success(w, http.StatusCreated, data)
}

func Error(w http.ResponseWriter, status int, code int, message string) {
	writeJSON(w, status, Response{
		Code:       code,
		Message:    message,
		Data:       nil,
		ServerTime: time.Now().Format(time.RFC3339),
	})
}

func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, CodeInvalidParams, message)
}

func Conflict(w http.ResponseWriter, code int, message string) {
	Error(w, http.StatusConflict, code, message)
}

func writeJSON(w http.ResponseWriter, status int, payload Response) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
