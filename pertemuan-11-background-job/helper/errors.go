package helper

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

const (
	CodeValidation           = "VALIDATION_ERROR"
	CodeBadRequest           = "BAD_REQUEST"
	CodeUnauthorized         = "UNAUTHORIZED"
	CodeForbidden            = "FORBIDDEN"
	CodeNotFound             = "NOT_FOUND"
	CodeConflict             = "CONFLICT"
	CodeUnsupportedMediaType = "UNSUPPORTED_MEDIA_TYPE"
	CodeNotAcceptable        = "NOT_ACCEPTABLE"
	CodeTooManyRequests      = "TOO_MANY_REQUESTS"
	CodeInternal             = "INTERNAL_ERROR"
)

type AppError struct {
	Status  int
	Code    string
	Message string
	Fields  map[string]string
	cause   error
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.cause }
func (e *AppError) Cause() error  { return e.cause }

func codeForStatus(status int) string {
	codes := map[int]string{400: CodeBadRequest, 401: CodeUnauthorized, 403: CodeForbidden, 404: CodeNotFound, 409: CodeConflict, 406: CodeNotAcceptable, 415: CodeUnsupportedMediaType, 422: CodeValidation, 429: CodeTooManyRequests}
	if code, ok := codes[status]; ok {
		return code
	}
	return CodeInternal
}

func NewAppError(status int, message string, details ...any) error {
	var fields map[string]string
	if len(details) > 0 {
		fields, _ = details[0].(map[string]string)
	}
	return &AppError{Status: status, Code: codeForStatus(status), Message: message, Fields: fields}
}

func Internal(err error) error {
	return &AppError{Status: 500, Code: CodeInternal, Message: "terjadi error pada server", cause: err}
}

func NewValidationError(fields map[string]string) error {
	return &AppError{Status: 422, Code: CodeValidation, Message: "validasi gagal", Fields: fields}
}

func EncodeCursor(createdAt time.Time, id int) (string, error) {
	if id < 1 || createdAt.IsZero() {
		return "", errors.New("invalid cursor values")
	}
	raw, err := json.Marshal(struct {
		CreatedAt time.Time `json:"created_at"`
		ID        int       `json:"id"`
	}{createdAt, id})
	if err != nil {
		return "", fmt.Errorf("encode cursor: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}
func DecodeCursor(value string) (time.Time, int, error) {
	raw, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return time.Time{}, 0, fmt.Errorf("invalid cursor: %w", err)
	}
	var cursor struct {
		CreatedAt time.Time `json:"created_at"`
		ID        int       `json:"id"`
	}
	if err := json.Unmarshal(raw, &cursor); err != nil || cursor.ID < 1 || cursor.CreatedAt.IsZero() {
		return time.Time{}, 0, errors.New("invalid cursor")
	}
	return cursor.CreatedAt, cursor.ID, nil
}
