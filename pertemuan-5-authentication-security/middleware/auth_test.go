package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"tugas1-go/pertemuan-5-authentication-security/app/model"
	"tugas1-go/pertemuan-5-authentication-security/helper"
)

const testSecret = "12345678901234567890123456789012"

func TestRequireAuth(t *testing.T) {
	manager := helper.NewJWTManager(testSecret, "test", time.Hour)
	valid, err := manager.GenerateAccessToken(model.User{ID: 9, Username: "tester", Role: "user"})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name, auth, message string
		status              int
	}{
		{"missing", "", "header Authorization tidak ada atau salah bentuk", 401},
		{"malformed", "Basic abc", "header Authorization tidak ada atau salah bentuk", 401},
		{"empty", "Bearer", "header Authorization tidak ada atau salah bentuk", 401},
		{"tampered", "Bearer " + valid + "x", "access token tidak valid", 401},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", RequireAuth(manager), func(c *fiber.Ctx) error { return c.SendStatus(204) })
			req := httptest.NewRequest("GET", "/", nil)
			if tc.auth != "" {
				req.Header.Set("Authorization", tc.auth)
			}
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			if resp.StatusCode != tc.status {
				t.Fatalf("got %d", resp.StatusCode)
			}
			if resp.Header.Get("WWW-Authenticate") != `Bearer realm="api"` {
				t.Fatal("missing challenge")
			}
		})
	}
	validApp := fiber.New()
	validApp.Get("/", RequireAuth(manager), func(c *fiber.Ctx) error {
		user, ok := helper.CurrentUser(c)
		if !ok || user.UserID != 9 {
			t.Fatal("identity missing")
		}
		return c.SendStatus(204)
	})
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "bearer "+valid)
	resp, err := validApp.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 204 {
		t.Fatalf("got %d", resp.StatusCode)
	}
}

func TestRequireAuthExpired(t *testing.T) {
	manager := helper.NewJWTManager(testSecret, "test", -time.Second)
	token, err := manager.GenerateAccessToken(model.User{ID: 1, Username: "x", Role: "user"})
	if err != nil {
		t.Fatal(err)
	}
	app := fiber.New()
	app.Get("/", RequireAuth(manager))
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 401 {
		t.Fatalf("got %d", resp.StatusCode)
	}
}

func TestLoginRateLimiter(t *testing.T) {
	app := fiber.New()
	app.Get("/", LoginRateLimiter(), func(c *fiber.Ctx) error { return c.SendStatus(204) })
	for i := 0; i < 5; i++ {
		resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 204 {
			t.Fatalf("request %d got %d", i+1, resp.StatusCode)
		}
	}
	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 429 {
		t.Fatalf("got %d", resp.StatusCode)
	}
	if resp.Header.Get("Retry-After") != "60" {
		t.Fatal("missing Retry-After")
	}
}
