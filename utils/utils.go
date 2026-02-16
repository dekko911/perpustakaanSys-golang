// INGAT LOG ERROR NYA KONTOL, PADAHAL BISA LOG ERROR LO UNTUK CEK SEGALA MASALAH, LITERASI / ATAU DIBACA DONG ANJING
package utils

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"path"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/go-playground/locales/id"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/perpus_backend/config"
	"github.com/redis/go-redis/v9"
	"github.com/rs/xid"

	"github.com/bytedance/sonic"
	"github.com/meilisearch/meilisearch-go"
)

var (
	NewMSClient = meilisearch.New(config.Env.MeilisearchURL,
		meilisearch.WithAPIKey(config.Env.MSApiKey),
		meilisearch.WithCustomJsonMarshaler(sonic.Marshal),
		meilisearch.WithCustomJsonUnmarshaler(sonic.Unmarshal),
		meilisearch.WithContentEncoding(meilisearch.GzipEncoding, meilisearch.BestSpeed))

	newIndonesiaLocale       = id.New()
	newUniTrans              = ut.New(newIndonesiaLocale, newIndonesiaLocale)
	GetTranslateIndonesia, _ = newUniTrans.GetTranslator("id_ID")

	bucketR2 = config.Env.CFBucketName

	NewValidate = validator.New(validator.WithRequiredStructEnabled()) // validate the request input.
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

// payload input must be pointed to payload param.
func ParseJSON(r *http.Request, payload any) error {
	if r.Body == http.NoBody {
		return errors.New("missing request body")
	}

	cfg := sonic.Config{
		CaseSensitive:    true,
		CompactMarshaler: true,
	}

	return cfg.Froze().NewDecoder(r.Body).Decode(payload)
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

	return cfg.Froze().NewEncoder(w).Encode(d)
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

// SetRedisKeyForPagination returns a Redis key for paginated data in the form
// "<keyName>:p:<page>:l:<limit>".
// It validates inputs and returns an error if validation fails.
// - keyName must be non-empty.
// - page must be >= 1.
// - limit must be >= 10.
// On success it returns the formatted key string and a nil error; otherwise an error describing the validation failure.
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
	if len(name) < 1 {
		return false // error
	}

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
	if len(roles) < 1 && len(targetRoles) < 1 {
		log.Printf("the roles or target roles is empty, see: roles: %v, target roles: %v", roles, targetRoles)
		return false // error
	}

	slices.Sort(roles)       // sort to ascending
	slices.Sort(targetRoles) // sort to ascending

	// if slices not sorted yet, throw error and just set to false.
	if !slices.IsSorted(roles) && !slices.IsSorted(targetRoles) {
		log.Println("the roles or target roles isn't sorted")
		return false
	}

	// brute force algorithm
	// or = outer loop, ir = inner loop
	for or := range roles {
		for ir := range targetRoles {
			if roles[or] != "" && targetRoles[ir] != "" {
				if roles[or] == targetRoles[ir] {
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

// TransformValidationErrorsWithLangIndonesian converts validation errors to a map of field names and their corresponding
// Indonesian error messages. Each key and value in the returned map is converted to lowercase.
// It takes validator.ValidationErrors as input and returns a map where keys are field names and values are translated error messages.
func TransformValidationErrorsWithLangIndonesian(vErrors validator.ValidationErrors) map[string]string {
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

// GetKeyFilepath constructs and returns the full URL path for a key file based on the environment configuration.
// It takes the keyFilepath (relative path to the key file) and a boolean isPrivate indicating whether the file is private.
// In production, the function requires the port to be empty and returns an error if it's not.
// The returned URL will include either "private" or "public" in the path depending on the isPrivate flag.
// In non-production environments, the port is included in the URL.
// Returns the constructed URL string and an error if any.
func GetKeyFilepath(keyFilepath string, isPrivate bool) (string, error) {
	if config.Env.AppENV == "production" {

		if config.Env.Port == "" {

			if isPrivate {
				return fmt.Sprintf("%s/private/%s", config.Env.AppURL, keyFilepath), nil
			}

			return fmt.Sprintf("%s/public/%s", config.Env.AppURL, keyFilepath), nil
		}

		return "-", errors.New("the port must be empty!")
	}

	if isPrivate {
		return fmt.Sprintf("%s:%s/private/%s", config.Env.AppURL, config.Env.Port, keyFilepath), nil
	}

	return fmt.Sprintf("%s:%s/public/%s", config.Env.AppURL, config.Env.Port, keyFilepath), nil
}

func R2GetObject(ctx context.Context, keyFilepath string) (*s3.GetObjectOutput, string, error) {
	if len(keyFilepath) < 1 {
		return nil, "", errors.New("you must fill keyFilepath first")
	}

	ext := path.Ext(keyFilepath)

	// init r2 storage
	clientR2, err := config.R2Storage(ctx)
	if err != nil {

		return nil, "", err
	}

	output, err := clientR2.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &config.Env.CFBucketName,
		Key:    &keyFilepath,
	})
	if err != nil {
		var noKey *types.NoSuchKey

		if errors.As(err, &noKey) {

			log.Printf("Can't get object %s from bucket %s. No such key exists.\n", keyFilepath, config.Env.CFBucketName)

			err = noKey
		} else {

			log.Printf("Couldn't get object %v:%v. Here's why: %v\n", config.Env.CFBucketName, keyFilepath, err)
		}

		return nil, "", err
	}

	return output, ext, nil
}

// fileMode = original or random.
func SetNewFilenameImg(ctx context.Context, filenameMode string, headerSrcFile *multipart.FileHeader, directoryPath string) (string, error) {

	if len(directoryPath) < 1 {

		return "-", errors.New("you must put the path directory")
	}

	if filenameMode != "original" && filenameMode != "random" {

		return "-", errors.New("invalid filename mode, only two: original & random")
	}

	// init multipart.File
	srcFile, err := headerSrcFile.Open()
	if err != nil {
		return "-", err
	}

	defer srcFile.Close()

	// init filename original
	newFilename := headerSrcFile.Filename

	fileExt := filepath.Ext(newFilename) // init extension fileImg

	if fileExt != ".png" && fileExt != ".jpg" && fileExt != ".jpeg" {

		return "-", fmt.Errorf("only support file '.jpg, .jpeg, and .png'. current file extension: %s", fileExt)
	}

	if headerSrcFile.Size <= 1<<20 {

		uniqueString := xid.New().String() // set the new unique filename

		switch filenameMode {
		case "random":
			newFilename = uniqueString + fileExt // set the new filename
		case "original":
			newFilename = strings.ReplaceAll(newFilename, " ", "-")
		}

		// init r2 storage
		clientR2, err := config.R2Storage(ctx)
		if err != nil {

			return "-", err
		}

		// init filepath for r2
		keyFilepath := path.Join(directoryPath, newFilename)

		_, err = clientR2.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      &bucketR2,
			Body:        srcFile,
			Key:         &keyFilepath,
			ContentType: aws.String(headerSrcFile.Header.Get("Content-Type")),
		})
		if err != nil {
			var apiErr smithy.APIError

			if errors.As(err, &apiErr) && apiErr.ErrorCode() == "EntityTooLarge" {
				// rewrite the var err
				err = fmt.Errorf("Error while uploading object to %s. The object is too large.\n"+
					"To upload objects larger than 5GB, use multipart upload API (5TB max).",
					config.Env.CFBucketName)

				return "-", err
			}

			// rewrite the var err
			err = fmt.Errorf("Couldn't upload file %v to %v:%v. Here's why: %v\n",
				newFilename, config.Env.CFBucketName, directoryPath, err)

			return "-", err
		}

		return keyFilepath, nil // <- this was right, below this was false case
	}

	return "-", errors.New("only serve file under 1mb")
}

// fileMode = original or random.
func UpdateTheFilenameImg(ctx context.Context, filenameMode string, headerSrcFile *multipart.FileHeader, directoryPath, oldKeyFileImgPath string) (string, error) {

	if len(oldKeyFileImgPath) < 1 {

		return "-", errors.New("you must put the oldKeyFilepath")
	} else if len(directoryPath) < 1 {

		return oldKeyFileImgPath, errors.New("you must put the path directory")
	}

	if filenameMode != "original" && filenameMode != "random" {

		return oldKeyFileImgPath, errors.New("invalid filename mode, only two: original & random")
	}

	srcFile, err := headerSrcFile.Open()
	if err != nil {

		return oldKeyFileImgPath, err
	}

	defer srcFile.Close()

	newFilename := headerSrcFile.Filename // set the filename

	fileExt := filepath.Ext(newFilename) // init extension filename

	if fileExt != ".png" && fileExt != ".jpg" && fileExt != ".jpeg" {

		return oldKeyFileImgPath, fmt.Errorf("only support file '.jpg, .jpeg, and .png'. current file extension: %s", fileExt)
	}

	if headerSrcFile.Size <= 1<<20 {

		if oldKeyFileImgPath != "-" {
			if err := DeleteFilepathWithFilename(ctx, oldKeyFileImgPath); err != nil {

				return oldKeyFileImgPath, err
			}
		}

		uniqueString := xid.New().String() // set the new unique filename

		switch filenameMode {
		case "random":
			newFilename = uniqueString + fileExt // set the new filename
		case "original":
			newFilename = strings.ReplaceAll(newFilename, " ", "-")
		}

		newKeyFilepath := path.Join(directoryPath, newFilename) // set the filename + dirPath

		clientR2, err := config.R2Storage(ctx)
		if err != nil {

			return "-", err
		}

		_, err = clientR2.PutObject(ctx, &s3.PutObjectInput{
			Bucket:      &bucketR2,
			Body:        srcFile,
			Key:         &newKeyFilepath,
			ContentType: aws.String(headerSrcFile.Header.Get("Content-Type")),
		})
		if err != nil {
			var apiErr smithy.APIError

			if errors.As(err, &apiErr) && apiErr.ErrorCode() == "EntityTooLarge" {
				// rewrite the var err
				err = fmt.Errorf("Error while uploading object to %s. The object is too large.\n"+
					"To upload objects larger than 5GB, use the S3 console (160GB max)\n"+
					"or the multipart upload API (5TB max).", config.Env.CFBucketName)

				return "-", err
			}

			// rewrite the var err
			err = fmt.Errorf("Couldn't upload file %v to %v:%v. Here's why: %v\n",
				newFilename, config.Env.CFBucketName, directoryPath, err)

			return "-", err
		}

		return newKeyFilepath, nil // <- this was right, below this was false case
	}

	return oldKeyFileImgPath, errors.New("only serve file under 1mb")
}

func SetOriginalFilenamePDF(ctx context.Context, headerSrcFilePDF *multipart.FileHeader, directoryPath string) (string, error) {

	if len(directoryPath) < 1 {

		return "-", errors.New("you must put the path directory")
	}

	srcFilePDF, err := headerSrcFilePDF.Open()
	if err != nil {
		return "-", err
	}

	defer srcFilePDF.Close()

	newFilenamePDF := headerSrcFilePDF.Filename

	fileExt := filepath.Ext(newFilenamePDF)

	if fileExt == ".pdf" {
		if headerSrcFilePDF.Size <= 8<<20 {

			newFilenamePDF = strings.ReplaceAll(newFilenamePDF, " ", "-")

			keyFilePath := path.Join(directoryPath, newFilenamePDF)

			clientR2, err := config.R2Storage(ctx)
			if err != nil {

				return "-", err
			}

			_, err = clientR2.PutObject(ctx, &s3.PutObjectInput{
				Bucket:      &bucketR2,
				Key:         &keyFilePath,
				Body:        srcFilePDF,
				ContentType: aws.String(headerSrcFilePDF.Header.Get("Content-Type")),
			})
			if err != nil {
				var apiErr smithy.APIError

				if errors.As(err, &apiErr) && apiErr.ErrorCode() == "EntityTooLarge" {
					// rewrite the var err
					err = fmt.Errorf("Error while uploading object to %s. The object is too large.\n"+
						"To upload objects larger than 5GB, use the S3 console (160GB max)\n"+
						"or the multipart upload API (5TB max).", config.Env.CFBucketName)

					return "-", err
				}

				// rewrite the var err
				err = fmt.Errorf("Couldn't upload file %v to %v:%v. Here's why: %v\n",
					newFilenamePDF, config.Env.CFBucketName, directoryPath, err)

				return "-", err
			}

			return keyFilePath, nil
		}

		return "-", errors.New("only serve file PDF under 8mb")
	}

	return "-", fmt.Errorf("only support file '.pdf'. current file extension: %s", fileExt)
}

func UpdateTheOriginalFilenamePDF(ctx context.Context, headerSrcFilePDF *multipart.FileHeader, directoryPath, oldKeyFilePDFPath string) (string, error) {

	if len(oldKeyFilePDFPath) < 1 {

		return "-", errors.New("you must put the oldKeyFilepath")
	} else if len(directoryPath) < 1 {

		return oldKeyFilePDFPath, errors.New("you must put the path directory")
	}

	srcFilePDF, err := headerSrcFilePDF.Open()
	if err != nil {

		return oldKeyFilePDFPath, err
	}

	newFilePDF := headerSrcFilePDF.Filename

	fileExt := filepath.Ext(newFilePDF)

	if fileExt == ".pdf" {
		if headerSrcFilePDF.Size <= 8<<20 {

			if oldKeyFilePDFPath != "-" {
				if err := DeleteFilepathWithFilename(ctx, oldKeyFilePDFPath); err != nil {

					return oldKeyFilePDFPath, err
				}
			}

			newFilePDF = strings.ReplaceAll(newFilePDF, " ", "-")

			keyFilepath := path.Join(directoryPath, newFilePDF)

			clientR2, err := config.R2Storage(ctx)
			if err != nil {

				return "-", err
			}

			_, err = clientR2.PutObject(ctx, &s3.PutObjectInput{
				Bucket:      &bucketR2,
				Key:         &keyFilepath,
				Body:        srcFilePDF,
				ContentType: aws.String(headerSrcFilePDF.Header.Get("Content-Type")),
			})
			if err != nil {
				var apiErr smithy.APIError

				if errors.As(err, &apiErr) && apiErr.ErrorCode() == "EntityTooLarge" {
					// rewrite the var err
					err = fmt.Errorf("Error while uploading object to %s. The object is too large.\n"+
						"To upload objects larger than 5GB, use the S3 console (160GB max)\n"+
						"or the multipart upload API (5TB max).", config.Env.CFBucketName)

					return "-", err
				}

				// rewrite the var err
				err = fmt.Errorf("Couldn't upload file %v to %v:%v. Here's why: %v\n",
					newFilePDF, config.Env.CFBucketName, directoryPath, err)

				return "-", err
			}

			return newFilePDF, nil
		}

		return oldKeyFilePDFPath, errors.New("only serve file PDF under 8mb")
	}

	return oldKeyFilePDFPath, fmt.Errorf("only support file '.pdf'. current file extension: %s", fileExt)
}

// DeleteFilepathWithFilename deletes one or more objects from R2 storage by their file paths.
// It takes a context and one or more S3 key filepaths as input.
// Returns an error if the filepaths slice is empty, if R2 storage connection fails,
// if the bucket does not exist, or if any deletion errors occur.
func DeleteFilepathWithFilename(ctx context.Context, keyFilepaths ...string) error {

	if len(keyFilepaths) < 1 {
		return fmt.Errorf("the target %v is empty", keyFilepaths)
	}

	clientR2, err := config.R2Storage(ctx)
	if err != nil {
		return err
	}

	newDelObjs := make([]types.ObjectIdentifier, 0, len(keyFilepaths))

	for i := range keyFilepaths {

		newDelObjs = append(newDelObjs, types.ObjectIdentifier{Key: &keyFilepaths[i]}) // iterasi terus sampai value dari var i habis
	}

	delOutput, err := clientR2.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: &bucketR2,
		Delete: &types.Delete{
			Objects: newDelObjs,
			Quiet:   aws.Bool(true),
		},
	})

	if err != nil || len(delOutput.Errors) > 0 {
		if err != nil {
			var noBucket *types.NoSuchBucket // check aja, siapa tau kena invalid mem addr

			if errors.As(err, &noBucket) {
				// rewrite err
				err = fmt.Errorf("Bucket %s does not exist.\n", bucketR2)

				return err
			}

			return err

		} else if len(delOutput.Errors) > 0 {
			// rewrite err
			err = fmt.Errorf("%s", *delOutput.Errors[0].Message) // rewrite variable err

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

func IsIndexMeiliAvailable(ctx context.Context, client meilisearch.ServiceManager, index string) bool {
	if len(index) < 1 {
		return false
	}

	if _, err := client.GetIndexWithContext(ctx, index); err != nil {
		return false
	}

	return true
}
