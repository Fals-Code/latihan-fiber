package helper

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type AppError struct {
	Status  int
	Message string
	Errors  any
	Err     error
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }
func NewAppError(status int, message string, details ...any) error {
	var errs any
	if len(details) > 0 {
		errs = details[0]
	}
	return &AppError{Status: status, Message: message, Errors: errs}
}
func NewValidationError(errs map[string]string) error {
	return &AppError{Status: 422, Message: "validasi gagal", Errors: errs}
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
