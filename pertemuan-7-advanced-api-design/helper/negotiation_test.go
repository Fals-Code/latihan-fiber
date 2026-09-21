package helper

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"tugas1-go/pertemuan-7-advanced-api-design/app/model"
)

func TestNegotiateRejectsUnsupportedAccept(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: func(c *fiber.Ctx, err error) error {
		appErr, ok := err.(*AppError)
		if !ok {
			return err
		}
		return c.SendStatus(appErr.Status)
	}})
	app.Get("/", func(c *fiber.Ctx) error { return Negotiate(c, 200, "ok", nil, nil) })
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "text/plain")
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != fiber.StatusNotAcceptable {
		t.Fatalf("expected 406, got status=%d err=%v", resp.StatusCode, err)
	}
}

func TestNegotiateCSV(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error { return Negotiate(c, 200, "ok", []model.Student{{ID: 1, Name: "A"}}, nil) })
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "text/csv")
	resp, err := app.Test(req)
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("request failed: %v", err)
	}
	if got := resp.Header.Get("Content-Type"); got == "" {
		t.Fatal("missing content type")
	}
}
