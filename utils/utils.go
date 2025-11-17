package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"perpus_backend/config"
	"runtime"
	"strconv"
	"strings"
	"time"

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

// returned information into json type.
func WriteJSON(w http.ResponseWriter, statusCode int, d JsonData) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	return json.NewEncoder(w).Encode(d)
}

// returned information into json error type.
func WriteJSONError(w http.ResponseWriter, statusCode int, err error) {
	_, file, line, _ := runtime.Caller(1)

	switch config.Env.AppENV {
	case "production":
		WriteJSON(w, statusCode, JsonData{
			Code:   statusCode,
			Status: http.StatusText(statusCode),
		})
	case "debug":
		WriteJSON(w, statusCode, JsonData{
			Code:   statusCode,
			Error:  err.Error(),
			File:   file,
			Line:   line,
			Status: http.StatusText(statusCode),
		})
	default:
	}
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

func IsItInBaseDir(path, baseDir string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		return false
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false
		}

		return false
	}

	if info.IsDir() {
		return false
	}

	return len(absPath) >= len(absBaseDir) && absPath[:len(absBaseDir)] == absBaseDir
}

func ParseStringToInt(val string) int {
	n, _ := strconv.Atoi(val)

	return n
}

func ParseStringToFloat(val string) float64 {
	f, _ := strconv.ParseFloat(val, 64)

	return f
}

func ParseDateFromFormInput(inputDate string) time.Time {
	d, _ := time.Parse(time.DateOnly, inputDate)

	return d
}

// this was support names: admin, staff, user, guest, etc. out of that, is should be invalid.
func IsInputRoleNameWasValid(name string) bool {
	switch name {
	// admin
	case "admin":

		// staff
	case "staff":

		// user
	case "user":

		// guest
	case "guest":

		// viewer
	case "viewer":

		// editor
	case "editor":

		// manager
	case "manager":

		// author
	case "author":

		// owner
	case "owner":

		// developer
	case "developer":

		// operator
	case "operator":

		// auditor
	case "auditor":
		return false
	default:
		return true
	}

	return true
}
