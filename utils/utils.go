// INGAT LOG NYA KONTOL, PADAHAL ADA LOG LO UNTUK CEK SEGALA MASALAH, ANJING LO
package utils

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/locales/id"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/perpus_backend/config"
	"github.com/redis/go-redis/v9"
	"github.com/rs/xid"

	"github.com/bytedance/sonic"
	"github.com/gorilla/websocket"
	"github.com/meilisearch/meilisearch-go"
)

var (
	NewMSClient = meilisearch.New(config.Env.MeilisearchURL,
		meilisearch.WithAPIKey(config.Env.MSApiKey),
		meilisearch.WithCustomJsonMarshaler(sonic.Marshal),
		meilisearch.WithCustomJsonUnmarshaler(sonic.Unmarshal),
		meilisearch.WithContentEncoding(meilisearch.GzipEncoding, meilisearch.BestCompression))

	newIndonesiaLocale       = id.New()
	newUniTrans              = ut.New(newIndonesiaLocale, newIndonesiaLocale)
	GetTranslateIndonesia, _ = newUniTrans.GetTranslator("id_ID")

	NewValidate = validator.New(validator.WithRequiredStructEnabled()) // validate the request input.

	NewWSUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

// JsonResponse represents a standardized JSON payload used for HTTP responses.
//
// It centralizes common response elements so handlers can return a consistent
// structure. Fields are omitted from the JSON output when they are empty.
//
// Fields:
//
//	Data     - arbitrary response payload (any type, omitted if nil).
//	Error    - error details or object (omitted if nil).
//	Token    - authentication token to return to the client.
//	File     - file path related to the response.
//	Message  - status message manual input from user.
//	Status   - short status label (e.g., "success", "error").
//	Page     - current page number for paginated results.
//	LastPage - total number of pages available (useful for pagination).
//	Line     - auxiliary numeric value (e.g., line index).
//	Code     - numeric status code from response status.
type JsonResponse struct {
	Data  any `json:"data,omitempty"`
	Error any `json:"error,omitempty"`

	Token   string `json:"token,omitempty"`
	File    string `json:"file,omitempty"`
	Message string `json:"message,omitempty"`
	Status  string `json:"status,omitempty"`

	Page     int   `json:"page,omitempty"`
	LastPage int64 `json:"last_page,omitempty"`
	Line     int   `json:"line,omitempty"`
	Code     int   `json:"code,omitempty"`
}

// IsTesting returns true if the code is being executed in a test environment,
// false otherwise. It checks if the test.v flag is set, which is present during
// test execution via `go test`.
func IsTesting() bool {
	return flag.Lookup("test.v") != nil
}

// WriteJSON writes the provided JsonResponse to w as JSON.
// It sets the "Content-Type" header to "application/json" and writes the
// provided HTTP status code before encoding the body.
// The encoder is created from a sonic.Config with CaseSensitive, DisallowUnknownFields
// and SortMapKeys enabled; the config is frozen and used to encode the value.
// Encoding is performed on &d and any error from the encoder is returned.
// Note: headers and status code are sent before encoding, so callers should ensure
// the value is JSON-serializable because the response will already be committed
// even if encoding fails.
func WriteJSON(w http.ResponseWriter, statusCode int, d JsonResponse) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	cfg := sonic.Config{
		CaseSensitive:         true,
		DisallowUnknownFields: true,
		SortMapKeys:           true,
	}

	// encoder = Go -> change format -> JSON
	// decoder = JSON -> change format -> Go

	return cfg.Froze().NewEncoder(w).Encode(&d)
}

// WriteJSONError writes an error response in JSON format to the provided http.ResponseWriter.
// It accepts an HTTP status code and an arbitrary error value (err).
//
// Supported err types:
//   - error: treated as a standard error and its string is returned (in non-production modes).
//   - map[string]string: treated as structured error details.
//
// The function captures the calling file and line (via runtime.Caller) to include in debug/testing
// responses. Behavior depends on config.Env.AppENV:
//   - "production": only Code and Status are returned (no error details).
//   - "debug": returns Code, Status, Error (string or map), File and Line.
//   - other: if IsTesting() is true, behaves like "debug"; otherwise the process exits via log.Fatalf.
//
// Unsupported err types are logged (log.Printf) and no JSON response is written for them.
// The JSON payload is produced using WriteJSON and follows the JsonResponse structure
// (fields: Code, Status, Error, File, Line).
func WriteJSONError(w http.ResponseWriter, statusCode int, err any) {
	_, file, line, _ := runtime.Caller(1)

	statsCodeToStatsText := http.StatusText(statusCode)

	switch randType := err.(type) {

	default:
		log.Printf("invalid error type: %v", randType)

	// case type error
	case error:

		switch config.Env.AppENV {
		case "production":
			WriteJSON(w, statusCode, JsonResponse{
				Code:   statusCode,
				Status: statsCodeToStatsText,
			})
		case "debug":
			WriteJSON(w, statusCode, JsonResponse{
				Code:   statusCode,
				Error:  randType.Error(),
				File:   file,
				Line:   line,
				Status: statsCodeToStatsText,
			})
		default:
			if IsTesting() {
				WriteJSON(w, statusCode, JsonResponse{
					Code:   statusCode,
					Error:  randType.Error(),
					File:   file,
					Line:   line,
					Status: statsCodeToStatsText,
				})
			} else {
				log.Fatalf("invalid value app_env: %s", config.Env.AppENV)
			}
		}

	// case type map[string]string
	case map[string]string:

		switch config.Env.AppENV {
		case "production":
			WriteJSON(w, statusCode, JsonResponse{
				Code:   statusCode,
				Status: statsCodeToStatsText,
			})
		case "debug":
			WriteJSON(w, statusCode, JsonResponse{
				Code:   statusCode,
				Error:  randType,
				File:   file,
				Line:   line,
				Status: statsCodeToStatsText,
			})
		default:
			if IsTesting() {
				WriteJSON(w, statusCode, JsonResponse{
					Code:   statusCode,
					Error:  randType,
					File:   file,
					Line:   line,
					Status: statsCodeToStatsText,
				})
			} else {
				log.Fatalf("invalid value app_env: %s", config.Env.AppENV)
			}
		}

	}
}

// SetRedisKey formats a Redis key by combining keyName and keyValue with a colon separator.
// It validates that both keyName and keyValue are non-empty strings.
// Returns the formatted key string in the format "keyName:keyValue" or an error if either parameter is empty.
// {resource:value}
func SetRedisKey(keyName, keyValue string) (string, error) {
	if keyName == "" || keyValue == "" {
		return "", errors.New("keyName AND keyValue must have value")
	}

	return fmt.Sprintf("%s:%s", keyName, keyValue), nil
}

func SetRedisKeyForPagination(keyName string, page, limit int) (string, error) {
	if page < 1 || limit < 10 {

		return "", errors.New("page must be over than or equal 1 int value and limit must be over than or equal 10 int value")
	} else if keyName == "" {

		return "", errors.New("keyName must have value")
	}

	// p = page, l = limit
	return fmt.Sprintf("%s:p:%d:l:%d", keyName, page, limit), nil
}

// GetTokenFromRequest extracts the JWT token from the Authorization header of an HTTP request.
// It expects the token to be in the format "Bearer <token>". The function removes the "Bearer " prefix
// and trims any leading or trailing whitespace. If a valid token is found, it returns the token string;
// otherwise, it returns an empty string.
func GetTokenFromRequest(r *http.Request) string {
	tokenHeader := r.Header.Get("Authorization")
	tokenString := strings.TrimPrefix(tokenHeader, "Bearer ")
	tokenString = strings.TrimSpace(tokenString)

	if tokenString != "" {
		return tokenString
	}

	return ""
}

// IsItInBaseDir reports whether the provided path refers to a non-directory file located inside the given baseDir.
// It first converts both path and baseDir to absolute paths; if either conversion fails or the path does not exist, it returns false.
// If the path refers to a directory, the function returns false.
// Containment is determined by a simple string prefix comparison of the absolute paths; symbolic links are not fully resolved and this method can yield false positives for similarly named paths (e.g. "/base/dir2" vs "/base/dir"). Use more robust checks (e.g. filepath.Rel and verifying path separators or resolving symlinks) if strict containment is required.
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

// ParseStringToInt converts a decimal string to an int.
// It is a thin wrapper around strconv.Atoi() that returns the parsed integer.
// If parsing fails, the error is discarded and the function returns 0.
func ParseStringToInt(number string) int {
	i, _ := strconv.Atoi(number)
	return i
}

// ParseStringToFloat parses s as a float64 using strconv.ParseFloat() with 64-bit precision.
// If parsing fails, the error is discarded and 0.0 is returned. Accepts decimal and scientific notation.
func ParseStringToFloat(number string) float64 {
	f, _ := strconv.ParseFloat(number, 64)
	return f
}

// ParseStringToFormatDate parses a date string in the "2006-01-02" (time.DateOnly) layout and returns the resulting time.Time.
// The input must be in YYYY-MM-DD format. On parse failure the error is ignored and the zero value of time.Time is returned.
// The returned time represents the parsed date at midnight (00:00:00) in UTC.
func ParseStringToFormatDate(date string) time.Time {
	d, _ := time.Parse(time.DateOnly, date)
	// return d.GoString()
	return d
}

// IsInputRoleNameWasValid reports whether the provided role name is one of the
// allowed role names ("admin", "staff", "user"). The check is case-sensitive.
// It returns true when the name exactly matches an allowed role, otherwise false.
func IsInputRoleNameWasValid(name string) bool {
	roleNamesMap := map[string]struct{}{
		"admin": {},
		"staff": {},
		"user":  {},
	}

	_, exist := roleNamesMap[name]
	return exist
}

// CompareRole compares two slices of role strings and returns true if any non-empty role
// in the roles slice matches any non-empty role in the targetRoles slice.
// Both slices are sorted in ascending order before comparison.
func CompareRole(roles, targetRoles []string) bool {
	slices.Sort(roles)       // sort to ascending
	slices.Sort(targetRoles) // sort to ascending

	// brute force algorithm
	for _, s := range roles {
		for _, t := range targetRoles {
			if s != "" && t != "" {
				if s == t {
					return true
				}
			}
		}
	}

	return false
}

// GenerateSpecificID generates a formatted ID string with a given prefix and zero-padded number.
// It takes a prefix string, a number to be incremented, and a width for zero-padding.
// Returns the formatted ID string or an error if prefix is empty or width is zero.
//
// Example: prefix -> ID, number -> 0 + 1 = 1, width -> 3 = 001 || 4 = 0001
func GenerateSpecificID(prefix string, number int, width int) (string, error) {
	if prefix == "" || width == 0 {
		return "", errors.New("input valid prefix AND width")
	}

	return fmt.Sprintf("%s%0*d", prefix, width, number+1), nil // prefix itu adalah awalan kata
}

// TransformValidationErrorsWithLangIndonesia converts validation errors to a map of field names and their corresponding
// Indonesian error messages. Each key and value in the returned map is converted to lowercase.
// It takes validator.ValidationErrors as input and returns a map where keys are field names and values are translated error messages.
func TransformValidationErrorsWithLangIndonesia(vErrors validator.ValidationErrors) map[string]string {
	newVErrors := make(map[string]string)

	for _, e := range vErrors {
		key := e.Field()
		value := e.Translate(GetTranslateIndonesia)

		// for lowercase strings
		lowerKey := strings.ToLower(key)
		lowerValue := strings.ToLower(value)

		newVErrors[lowerKey] = lowerValue
	}

	return newVErrors
}

// SetNewFilenameImg validates and saves an uploaded image file and returns the filename actually used.
//
// The function accepts a filenameMode ("original" or "random"), a multipart.File containing the file
// data, a directoryPath to write into, the original file name (fileImg) to derive the extension, and
// the sizeFile in bytes. It only accepts extensions .png, .jpg and .jpeg and enforces a maximum size
// of 1<<20 bytes (1 MiB).
//
// Behavior:
//   - If filenameMode == "random", the saved filename is xid.New().String() + file extension.
//   - If filenameMode == "original", the original fileImg name is used (subject to validation).
//   - The destination path is formed by concatenating directoryPath and the chosen filename; directoryPath
//     is expected to include the appropriate path separator if needed.
//   - The function creates (os.Create) the destination file (truncating any existing file) and copies the
//     contents from srcFile into it. It does not close srcFile; the caller is responsible for closing it.
//   - srcFile should be positioned at the start of the file (or at the intended read position) before calling.
//
// Return values:
//   - On success returns the filename written (without directoryPath) and a nil error.
//   - On failure returns "-" and a non-nil error explaining the reason (invalid filename mode, unsupported
//     extension, file too large, or I/O error while creating or copying the file).
func SetNewFilenameImg(filenameMode string, srcFile multipart.File, directoryPath, fileImg string, sizeFile int64) (string, error) {
	if filenameMode != "original" && filenameMode != "random" {

		fileImg = "-" // for explicit things

		return fileImg, errors.New("invalid filename mode, only two: original & random")
	}

	fileExt := filepath.Ext(fileImg) // init extension fileImg

	if fileExt != ".png" && fileExt != ".jpg" && fileExt != ".jpeg" {

		fileImg = "-" // for explicit things

		return fileImg, fmt.Errorf("only support file '.jpg, .jpeg, and .png'. current file extension: %s", fileExt)
	}

	if sizeFile <= 1<<20 {

		uniqueString := xid.New().String() // set the new unique filename

		if filenameMode == "random" {
			fileImg = uniqueString + fileExt // set the new filename
		}

		filePath := filepath.Join(directoryPath, fileImg) // set the filename + dirPath

		cleanPath := filepath.Clean(filePath) // clean the path for not messing the path?

		destination, _ := os.Create(cleanPath)
		defer destination.Close()

		io.Copy(destination, srcFile)

		return fileImg, nil // <- this was right, below this was false case
	}

	fileImg = "-" // for explicit things

	return fileImg, errors.New("only serve file under 1mb")
}

// UpdateTheFilenameImg handles uploading and storing an image file with optional filename randomization.
// It validates the filename mode, file extension, and file size before saving the image to disk.
// If an old file exists at the specified path, it will be deleted before the new file is saved.
//
// Parameters:
//   - filenameMode: The naming strategy for the saved file, either "original" (keeps newFileImg name) or "random" (generates unique name)
//   - srcFile: The multipart.File to be uploaded
//   - directoryPath: The directory path where the image will be saved (should end with path separator)
//   - oldFileImg: The filename of the previous image to be deleted if it exists
//   - newFileImg: The desired filename (or base filename if using random mode); must include extension
//   - sizeFile: The size of the file in bytes; must be <= 1MB (1048576 bytes)
//
// Returns:
//   - string: The saved filename (either newFileImg or a randomly generated name with original extension)
//   - error: An error if filename mode is invalid, file extension is unsupported, file size exceeds limit, or oldFileImg file deletion fails
//
// Supported file extensions: .png, .jpg, .jpeg
// Maximum file size: 1MB
func UpdateTheFilenameImg(filenameMode string, srcFile multipart.File, directoryPath, oldFileImg, newFileImg string, sizeFile int64) (string, error) {
	if filenameMode != "original" && filenameMode != "random" {

		return oldFileImg, errors.New("invalid filename mode, only two: original & random")
	}

	fileExt := filepath.Ext(newFileImg) // init extension filename

	if fileExt != ".png" && fileExt != ".jpg" && fileExt != ".jpeg" {

		return oldFileImg, fmt.Errorf("only support file '.jpg, .jpeg, and .png'. current file extension: %s", fileExt)
	}

	if sizeFile <= 1<<20 {

		oldFilepath := filepath.Join(directoryPath, oldFileImg) // set filename here for make sure get the first the filename in param, you know what i mean

		if err := DeleteFilepathWithFilename(oldFilepath); err != nil {
			return oldFileImg, err
		}

		uniqueString := xid.New().String() // set the new unique filename

		if filenameMode == "random" {
			newFileImg = uniqueString + fileExt // set the new filename
		}

		newFilepath := filepath.Join(directoryPath, newFileImg) // set the filename + dirPath

		cleanPath := filepath.Clean(newFilepath)

		// os.Rename(newFilepath, oldFilepath) // set replace the oldFile to newFile

		destination, _ := os.Create(cleanPath)
		defer destination.Close()

		io.Copy(destination, srcFile)

		return newFileImg, nil // <- this was right, below this was false case
	}

	return oldFileImg, errors.New("only serve file under 1mb")
}

// SetOriginalFilenamePDF saves an uploaded PDF file to the specified directory if it meets size and format requirements.
// It validates that the file has a .pdf extension and does not exceed 8MB in size.
// If validation passes, the file is written to disk at the specified local directory path.
// srcFilePDF is the multipart file to be saved.
// directoryPath is the destination local directory path where the file will be stored.
// filePDF is the original filePDF of the uploaded file (header.Filename).
// sizeFile is the size of the file in bytes.
// It returns the filename if successful, or "-" with an error if validation fails.
// Possible errors include file size exceeding 8MB or unsupported file format.
func SetOriginalFilenamePDF(srcFilePDF multipart.File, directoryPath, filePDF string, sizeFile int64) (string, error) {
	fileExt := filepath.Ext(filePDF)

	if fileExt == ".pdf" {
		if sizeFile <= 8<<20 {

			filePath := filepath.Join(directoryPath, filePDF)

			cleanPath := filepath.Clean(filePath)

			destination, _ := os.Create(cleanPath)
			defer destination.Close()

			io.Copy(destination, srcFilePDF)

			return filePDF, nil
		}

		filePDF = "-"

		return filePDF, errors.New("only serve file PDF under 8mb")
	}

	filePDF = "-"

	return filePDF, fmt.Errorf("only support file '.pdf'. current file extension: %s", fileExt)
}

// UpdateTheOriginalFilenamePDF updates a PDF file on disk by replacing an existing file with a newly uploaded PDF.
//
// UpdateTheOriginalFilenamePDF expects an open multipart.File (srcFilePDF), a directory path (directoryPath),
// the current filename to remove (oldFilePDF), the new filename to write (newFilePDF), and the reported size of
// the new file in bytes (sizeFile).
//
// Behavior:
//
// - Validates that newFilePDF has the ".pdf" extension and that sizeFile is less than or equal to 8 MB (8 << 20).
// - If validation fails, returns oldFilePDF and a descriptive error.
// - If validation succeeds, it attempts to remove the existing file at directoryPath+oldFilePDF if it exists and is not a directory.
// - It then creates a new file at directoryPath+newFilePDF and copies the contents of srcFilePDF into it.
// - On success it returns newFilePDF and nil error.
//
// Notes and caveats:
//   - Paths are constructed by simple string concatenation of directoryPath and the filename (no path joining or cleaning).
//   - The function treats the given sizeFile as the authority for size checking (it does not re-check the stream length).
//   - The implementation may ignore or propagate errors from file operations; callers should be aware of possible side effects
//     (partial files written, old file removed) if an error occurs during creation/copy.
//   - Only files with the exact ".pdf" extension are accepted (case-sensitive).
func UpdateTheOriginalFilenamePDF(srcFilePDF multipart.File, directoryPath, oldFilePDF, newFilePDF string, sizeFile int64) (string, error) {
	fileExt := filepath.Ext(newFilePDF)

	if fileExt == ".pdf" {
		if sizeFile <= 8<<20 {

			oldFilepath := filepath.Join(directoryPath, oldFilePDF)

			if err := DeleteFilepathWithFilename(oldFilepath); err != nil {
				return oldFilePDF, err
			}

			filePath := filepath.Join(directoryPath, newFilePDF)

			cleanPath := filepath.Clean(filePath)

			destination, _ := os.Create(cleanPath)
			defer destination.Close()

			io.Copy(destination, srcFilePDF) // <- untuk copy seluruh data (dari metadata file, dll) ke param destination, agar bisa disimpan layaknya file biasanya, tapi juga support streaming dalam ukuran file yang besar

			return newFilePDF, nil
		}

		return oldFilePDF, errors.New("only serve file PDF under 8mb")
	}

	return oldFilePDF, fmt.Errorf("only support file '.pdf'. current file extension: %s", fileExt)
}

// DeleteFilepathWithFilename resolves and deletes one or more files given by their paths.
// It evaluates symbolic links and cleans the resolved path before attempting removal.
//
// For each provided path:
//
// - If resolving symlinks or stat reports the path does not exist, the function returns nil (no error).
//
// - If the path exists and is not a directory, the file is removed.
//
// - If the path exists and is a directory, it is left untouched.
//
// The function returns the first non-nil error encountered (other than file-not-found) and short-circuits on that error.
func DeleteFilepathWithFilename(filepathWithFilenames ...string) error {
	for _, v := range filepathWithFilenames {
		// get actual path & actual target file
		realPath, err := filepath.EvalSymlinks(v)
		if err != nil {

			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}

			return err
		}

		// clean the path
		cleanPath := filepath.Clean(realPath)

		// delete the file
		info, err := os.Stat(cleanPath)
		if err == nil {

			if !info.IsDir() {
				os.Remove(cleanPath)
			}

		} else {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}

			return err
		}

	}

	return nil
}

// InvalidateAllKeysInCache invalidates (deletes) cached Redis keys for the application.
// It scans the Redis keyspace for specific patterns, collects matching keys, and deletes
// them in a pipeline.
//
// The function:
//   - Uses SCAN to iterate keys for these patterns: "user:*", "role:*", "member:*",
//     "circulation:*", "book:*", and pagination keys "users:p:*", "members:p:*",
//     "circulations:p:*", "books:p:*".
//   - Accumulates found keys, then issues DEL commands in a pipeline to remove them.
//   - Uses the provided context.Context (ctx) for cancellation/deadline propagation.
//
// Parameters:
//   - rdb: Redis client used to perform SCAN and pipeline DEL operations.
//   - ctx: Context for controlling cancellation and timeouts.
//
// Returns:
//   - error if any scanning iterator reports an error or if the pipeline Exec fails.
//   - nil if no error occurred (including the case where no matching keys were found).
//
// Notes:
//   - SCAN is incremental and safe for large keyspaces but does not provide atomic
//     snapshots; keys may be added/removed concurrently.
//   - Deletions are performed via a pipeline for efficiency but are not atomic.
func InvalidateAllKeysInCache(rdb *redis.Client, ctx context.Context) error {
	// init all ids iteration keys
	iterUserIds := rdb.Scan(ctx, 0, "user:*", 0).Iterator()
	iterRoleIds := rdb.Scan(ctx, 0, "role:*", 0).Iterator()
	iterMemberIds := rdb.Scan(ctx, 0, "member:*", 0).Iterator()
	iterCirculationIds := rdb.Scan(ctx, 0, "circulation:*", 0).Iterator()
	iterBookIds := rdb.Scan(ctx, 0, "book:*", 0).Iterator()

	// init all pagination iteration keys
	iterUsersPagination := rdb.Scan(ctx, 0, "users:p:*", 0).Iterator()
	iterMembersPagination := rdb.Scan(ctx, 0, "members:p:*", 0).Iterator()
	iterCirculationsPagination := rdb.Scan(ctx, 0, "circulations:p:*", 0).Iterator()
	iterBooksPagination := rdb.Scan(ctx, 0, "books:p:*", 0).Iterator()

	var keys []string // init keys

	// iter user ids
	for iterUserIds.Next(ctx) {
		keys = append(keys, iterUserIds.Val())
	}

	if err := iterUserIds.Err(); err != nil {
		return err
	}

	// iter users pagination
	for iterUsersPagination.Next(ctx) {
		keys = append(keys, iterUsersPagination.Val())
	}

	if err := iterUsersPagination.Err(); err != nil {
		return err
	}

	// iter role ids
	for iterRoleIds.Next(ctx) {
		keys = append(keys, iterRoleIds.Val())
	}

	if err := iterRoleIds.Err(); err != nil {
		return err
	}

	// iter member ids
	for iterMemberIds.Next(ctx) {
		keys = append(keys, iterMemberIds.Val())
	}

	if err := iterMemberIds.Err(); err != nil {
		return err
	}

	// iter members pagination
	for iterMembersPagination.Next(ctx) {
		keys = append(keys, iterMembersPagination.Val())
	}

	if err := iterMembersPagination.Err(); err != nil {
		return err
	}

	// iter circulation ids
	for iterCirculationIds.Next(ctx) {
		keys = append(keys, iterCirculationIds.Val())
	}

	if err := iterCirculationIds.Err(); err != nil {
		return err
	}

	// iter circulations pagination
	for iterCirculationsPagination.Next(ctx) {
		keys = append(keys, iterCirculationsPagination.Val())
	}

	if err := iterCirculationsPagination.Err(); err != nil {
		return err
	}

	// iter book ids
	for iterBookIds.Next(ctx) {
		keys = append(keys, iterBookIds.Val())
	}

	if err := iterBookIds.Err(); err != nil {
		return err
	}

	// iter books pagination
	for iterBooksPagination.Next(ctx) {
		keys = append(keys, iterBooksPagination.Val())
	}

	if err := iterBooksPagination.Err(); err != nil {
		return err
	}

	// pipeLine redis del
	if len(keys) > 0 {
		pipeLine := rdb.Pipeline()

		for _, v := range keys {
			pipeLine.Del(ctx, v)
		}

		_, err := pipeLine.Exec(ctx)
		return err
	}

	return nil
}
