package middleware

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"

	"tugas1-go/uts-pbl-siakad-mini/app/model"
	"tugas1-go/uts-pbl-siakad-mini/app/repository"
	"tugas1-go/uts-pbl-siakad-mini/app/service"
)

const AuthUserLocal = "authenticated_user"

type AccountLookup interface {
	FindActiveAccount(ctx context.Context, userID int64) (model.CurrentUser, error)
}

func RequireAuth(secret []byte, accounts AccountLookup) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get(fiber.HeaderAuthorization)
		parts := strings.Fields(header)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" || strings.Contains(parts[1], ",") {
			return authFailure(c)
		}
		identity, err := service.ParseToken(parts[1], secret)
		if err != nil {
			return authFailure(c)
		}
		user, err := accounts.FindActiveAccount(c.UserContext(), identity.UserID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return authFailure(c)
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"success": false, "message": "Terjadi kesalahan pada server"})
		}
		if user.Role != identity.Role {
			return authFailure(c)
		}
		c.Locals(AuthUserLocal, user)
		return c.Next()
	}
}

func authFailure(c *fiber.Ctx) error {
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"success": false, "message": "Token tidak valid atau akun tidak aktif"})
}

type failureWindow struct {
	count int
	start time.Time
}

type loginFailureLimiter struct {
	mu       sync.Mutex
	windows  map[string]failureWindow
	now      func() time.Time
	limit    int
	duration time.Duration
}

func NewLoginFailureLimiter() *loginFailureLimiter {
	return &loginFailureLimiter{windows: make(map[string]failureWindow), now: time.Now, limit: 5, duration: time.Minute}
}

func LoginFailureLimiter() fiber.Handler {
	return NewLoginFailureLimiter().Middleware()
}

func (l *loginFailureLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.IP()
		if l.blocked(key) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{"success": false, "message": "Terlalu banyak percobaan login"})
		}
		if err := c.Next(); err != nil {
			return err
		}
		if c.Response().StatusCode() == fiber.StatusUnauthorized {
			l.recordFailure(key)
		}
		return nil
	}
}

func (l *loginFailureLimiter) blocked(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	window, ok := l.windows[key]
	if !ok || l.now().Sub(window.start) >= l.duration {
		delete(l.windows, key)
		return false
	}
	return window.count >= l.limit
}

func (l *loginFailureLimiter) recordFailure(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	window, ok := l.windows[key]
	if !ok || now.Sub(window.start) >= l.duration {
		l.windows[key] = failureWindow{count: 1, start: now}
		return
	}
	window.count++
	l.windows[key] = window
}
