package utils

import (
	"encoding/json"
	"net/http"
	"runtime"
	"strings"

	"github.com/go-playground/validator/v10"
)

type JsonData struct {
	Code    int    `json:"code,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
	File    string `json:"file,omitempty"`
	Line    int    `json:"line,omitempty"`
	Message string `json:"message,omitempty"`
	Status  string `json:"status,omitempty"`
	Token   string `json:"token,omitempty"`
}

// validate the request input.
var Validate = validator.New()

// returning info into json type.
func WriteJSON(w http.ResponseWriter, statusCode int, d JsonData) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	return json.NewEncoder(w).Encode(d)
}

// returning info into json error type.
func WriteJSONError(w http.ResponseWriter, statusCode int, err error) {
	_, file, line, _ := runtime.Caller(1)

	WriteJSON(w, statusCode, JsonData{
		Code:   statusCode,
		Error:  err.Error(),
		File:   file,
		Line:   line,
		Status: "error",
	})
}

// get the token from headers.
func GetTokenFromRequest(r *http.Request) string {
	tokenHeader := r.Header.Get("Authorization")
	tokenString := strings.TrimPrefix(tokenHeader, "Bearer ")
	tokenString = strings.TrimSpace(tokenString)

	if tokenString != "" {
		return tokenString
	}

	return ""
}
