package utils

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"perpus_backend/config"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/bytedance/sonic"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/websocket"
	"github.com/meilisearch/meilisearch-go"
)

var (
	MSClient = meilisearch.New(config.Env.MeilisearchURL,
		meilisearch.WithAPIKey(config.Env.MSApiKey),
		meilisearch.WithCustomJsonMarshaler(sonic.Marshal),
		meilisearch.WithCustomJsonUnmarshaler(sonic.Unmarshal))

	WSUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	Validate = validator.New() // validate the request input.
)

type JsonData struct {
	Data any `json:"data,omitempty"`

	Token   string `json:"token,omitempty"`
	Error   string `json:"error,omitempty"`
	File    string `json:"file,omitempty"`
	Message string `json:"message,omitempty"`
	Status  string `json:"status,omitempty"`

	Line int `json:"line,omitempty"`
	Code int `json:"code,omitempty"`
}

// for check if has do some go test, it will return true.
func IsTesting() bool {
	return flag.Lookup("test.v") != nil
}

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
		if IsTesting() {
			WriteJSON(w, statusCode, JsonData{
				Code:   statusCode,
				Error:  err.Error(),
				File:   file,
				Line:   line,
				Status: http.StatusText(statusCode),
			})
		} else {
			log.Fatalf("invalid value app_env: %s", config.Env.AppENV)
		}
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

func ParseStringToInt(number string) int {
	i, _ := strconv.Atoi(number)
	return i
}

func ParseStringToFloat(number string) float64 {
	f, _ := strconv.ParseFloat(number, 64)
	return f
}

func ParseStringToFormatDate(date string) time.Time {
	d, _ := time.Parse(time.DateOnly, date)
	return d
}

// this was support names: admin, staff, and user. out of that, it should be invalid.
func IsInputRoleNameWasValid(name string) bool {
	validRoleName := map[string]struct{}{
		"admin": {},
		"staff": {},
		"user":  {},
	}

	_, ok := validRoleName[name]
	return ok
}

func IsValidSortColumn(column string) bool {
	validColumn := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	return validColumn.MatchString(column)
}

func IsValidSortOrder(sortOrder string) bool {
	validSortOrder := map[string]struct{}{
		"ASC":  {},
		"DESC": {},
	}

	_, ok := validSortOrder[sortOrder]
	return ok
}

func ParseSliceRolesToFilteredString(s []string) string {
	filtered := make([]string, 0, len(s))

	for _, v := range s { // ["admin" "staff" "user"]
		v = strings.TrimSpace(v) // "admin""staff""user"

		// check it, if there has empty string or null, it will be skipped
		if v != "" {
			filtered = append(filtered, v)
		}
	}

	return strings.Join(filtered, ", ") // "admin, staff, user"
}

func CompareRole(sliceRoles, targetRole string) bool {
	for s := range strings.SplitSeq(sliceRoles, ",") {
		for t := range strings.SplitSeq(targetRole, ",") {
			if s == t { // [admin staff] compare [staff user]
				return true
			}
		}
	}

	return false
}

func GenerateSpecificID(prefix string, number int, width int) string {
	return fmt.Sprintf("%s%0*d", prefix, width, number+1) // prefix itu adalah awalan kata
}
