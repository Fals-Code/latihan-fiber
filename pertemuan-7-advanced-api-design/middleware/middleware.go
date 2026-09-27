package middleware

import (
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"tugas1-go/pertemuan-7-advanced-api-design/helper"
)

// Register memasang middleware global.
// Urutan pemasangan penting karena middleware dijalankan berurutan.
func Register(app *fiber.App, logger *slog.Logger, allowedOrigins []string) {
	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(helmet.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: strings.Join(allowedOrigins, ","),
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))
	app.Use(RequestLogger(logger))
}

// RequestLogger mencatat setiap request HTTP.
func RequestLogger(logger *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()
		status := c.Response().StatusCode()
		var appErr *helper.AppError
		if errors.As(err, &appErr) {
			status = appErr.Status
		} else if err != nil {
			status = fiber.StatusInternalServerError
		}

		requestID, _ := c.Locals("requestid").(string)
		attrs := []any{
			slog.String("request_id", requestID),
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Int("status", status),
			slog.Duration("duration", time.Since(start)),
			slog.String("ip", c.IP()),
		}
		if user, ok := helper.CurrentUser(c); ok {
			attrs = append(attrs, slog.Int("user_id", user.UserID), slog.String("role", user.Role))
		}

		if status >= 400 && status < 500 {
			logger.Warn("request_rejected", attrs...)
		} else if status >= 500 {
			logger.Error("request_failed", attrs...)
		} else {
			logger.Info("http_request", attrs...)
		}

		return err
	}
}

var methodsWithBody = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

// RequireJSON memastikan request yang memiliki body menggunakan JSON.
func RequireJSON(c *fiber.Ctx) error {
	if methodsWithBody[c.Method()] {
		contentType := c.Get("Content-Type")

		if !strings.HasPrefix(contentType, fiber.MIMEApplicationJSON) {
			return helper.NewAppError(fiber.StatusUnsupportedMediaType, "Content-Type harus application/json")
		}
	}

	return c.Next()
}
