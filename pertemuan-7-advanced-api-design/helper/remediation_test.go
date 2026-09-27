package helper

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"tugas1-go/pertemuan-7-advanced-api-design/app/model"
)

func TestNewValidationErrorUsesContract(t *testing.T) {
	err := NewValidationError(map[string]string{"username": "wajib diisi"})
	appErr, ok := err.(*AppError)
	if !ok || appErr.Code != CodeValidation || appErr.Fields["username"] == "" {
		t.Fatalf("unexpected validation error: %#v", err)
	}
}

func TestInternalPreservesCauseWithoutChangingMessage(t *testing.T) {
	cause := errors.New("database password leaked")
	appErr := Internal(cause).(*AppError)
	if appErr.Code != CodeInternal || appErr.Error() == cause.Error() || !errors.Is(appErr, cause) {
		t.Fatalf("internal error contract failed: %#v", appErr)
	}
}

func TestNegotiateWildcardDefaultsToJSON(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error { return Negotiate(c, 200, "ok", map[string]string{"a": "b"}, nil) })
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "*/*")
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != 200 || resp.Header.Get("Content-Type") == "text/csv; charset=utf-8" {
		t.Fatalf("wildcard negotiation failed: status=%d err=%v", resp.StatusCode, err)
	}
}

func TestStudentValidationUsesJSONFieldAndCustomMessage(t *testing.T) {
	errs := ValidateRequest(struct {
		NIM int `json:"nim" validate:"required,studentnim"`
	}{})
	if errs["nim"] == "" {
		t.Fatalf("expected non-empty nim validation message: %#v", errs)
	}
}

func TestStrongPasswordPolicy(t *testing.T) {
	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{"strong", "Abcdef1!", true},
		{"too short", "abc", false},
		{"common", "password123", false},
		{"without digit", "tanpaangka", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateRequest(model.RegisterRequest{Username: "coba01", Email: "c@example.test", Password: tt.password})
			if (len(errs) == 0) != tt.valid {
				t.Fatalf("password validity mismatch: password=%q errors=%#v", tt.password, errs)
			}
		})
	}
}

func TestNegotiateCSVHeaders(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error { return Negotiate(c, 200, "ok", []model.Student{{ID: 1, Name: "A"}}, nil) })
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "text/csv")
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != 200 || resp.Header.Get("Content-Type") != "text/csv; charset=utf-8" || resp.Header.Get("Content-Disposition") == "" {
		t.Fatalf("csv negotiation failed: status=%d err=%v", resp.StatusCode, err)
	}
}
