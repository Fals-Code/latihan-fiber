package middleware

import (
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestLoginFailureLimiterAllowsFiveFailuresThenBlocksNext(t *testing.T) {
	app := fiber.New()
	app.Post("/login", LoginFailureLimiter(), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusUnauthorized)
	})

	for attempt := 1; attempt <= 5; attempt++ {
		response, err := app.Test(httptest.NewRequest("POST", "/login", nil), -1)
		if err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != fiber.StatusUnauthorized {
			t.Fatalf("attempt %d: expected 401, got %d", attempt, response.StatusCode)
		}
	}
	response, err := app.Test(httptest.NewRequest("POST", "/login", nil), -1)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("sixth attempt: expected 429, got %d", response.StatusCode)
	}
}

func TestLoginFailureLimiterConcurrentRequestsAreSafe(t *testing.T) {
	limiter := NewLoginFailureLimiter()
	app := fiber.New()
	app.Post("/login", limiter.Middleware(), func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusUnauthorized) })
	var wait sync.WaitGroup
	for i := 0; i < 40; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			_, _ = app.Test(httptest.NewRequest("POST", "/login", nil), -1)
		}()
	}
	wait.Wait()
	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	if got := limiter.windows["0.0.0.0"].count; got > 40 {
		t.Fatalf("unexpected failure count: %d", got)
	}
}

func TestLoginFailureLimiterResetsAfterWindow(t *testing.T) {
	limiter := NewLoginFailureLimiter()
	now := time.Now()
	limiter.now = func() time.Time { return now }
	for i := 0; i < 5; i++ {
		limiter.recordFailure("client")
	}
	if !limiter.blocked("client") {
		t.Fatal("expected client to be blocked")
	}
	now = now.Add(time.Minute)
	if limiter.blocked("client") {
		t.Fatal("expected expired window to unblock client")
	}
}
