package utils

import (
	"encoding/json"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

var (
	// fileName Rand Using Time.
	Filename = randFileNameUsingTime()

	// validate the request input.
	Validate = validator.New()
)

// returning info into json type.
func WriteJSON(w http.ResponseWriter, statusCode int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	return json.NewEncoder(w).Encode(v)
}

// returning info into json error type.
func WriteJSONError(w http.ResponseWriter, statusCode int, err error) {
	_, file, line, _ := runtime.Caller(1)

	WriteJSON(w, statusCode, map[string]any{
		"code":   statusCode,
		"error":  err.Error(),
		"file":   file,
		"line":   line,
		"status": "error",
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

// making rand fileName.
func randFileNameUsingTime() string {
	loc, err := time.LoadLocation("Asia/Kuala_Lumpur")
	if err != nil {
		return err.Error()
	}

	timeInt := time.Now().In(loc).Unix()
	time := strconv.Itoa(int(timeInt))

	return time
}
