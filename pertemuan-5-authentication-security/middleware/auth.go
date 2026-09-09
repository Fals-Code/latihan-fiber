package middleware

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"

	"tugas1-go/pertemuan-5-authentication-security/helper"
)

const bearerChallenge = `Bearer realm="api"`

func RequireAuth(jwtManager *helper.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderWWWAuthenticate, bearerChallenge)
		raw, err := bearerToken(c)
		if err != nil {
			return helper.Fail(c, fiber.StatusUnauthorized, "header Authorization tidak ada atau salah bentuk")
		}
		identity, err := jwtManager.ParseAccessToken(raw)
		if errors.Is(err, helper.ErrExpiredToken) {
			return helper.Fail(c, fiber.StatusUnauthorized, "access token kedaluwarsa")
		}
		if err != nil {
			return helper.Fail(c, fiber.StatusUnauthorized, "access token tidak valid")
		}
		c.Locals(helper.LocalsAuthUser, identity)
		return c.Next()
	}
}

func bearerToken(c *fiber.Ctx) (string, error) {
	parts := strings.Fields(c.Get(fiber.HeaderAuthorization))
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", errors.New("invalid bearer header")
	}
	return parts[1], nil
}

func LoginRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:          5,
		Expiration:   time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string { return c.IP() },
		LimitReached: func(c *fiber.Ctx) error {
			c.Set(fiber.HeaderRetryAfter, "60")
			return helper.Fail(c, fiber.StatusTooManyRequests, "terlalu banyak percobaan login, coba lagi dalam satu menit")
		},
	})
}
