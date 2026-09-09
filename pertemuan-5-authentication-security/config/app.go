package config

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"tugas1-go/pertemuan-5-authentication-security/helper"
	"tugas1-go/pertemuan-5-authentication-security/middleware"
	"tugas1-go/pertemuan-5-authentication-security/route"
)

// NewApp merakit aplikasi Fiber, middleware, dan route.
func NewApp(logger *slog.Logger, deps route.Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      GetEnv("APP_NAME", "Praktikum Backend Lanjut"),
		BodyLimit:    1 * 1024 * 1024,
		ErrorHandler: newErrorHandler(logger),
	})

	middleware.Register(app, logger, GetEnvList("ALLOWED_ORIGINS", "http://localhost:5173"))
	route.Register(app, deps)

	// Menangani endpoint yang tidak dikenal.
	app.Use(func(c *fiber.Ctx) error {
		return helper.Fail(
			c,
			fiber.StatusNotFound,
			"endpoint tidak ditemukan",
		)
	})

	return app
}

// newErrorHandler menjadi jaring pengaman terakhir untuk error
// yang belum ditangani oleh service.
func newErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		status := fiber.StatusInternalServerError
		message := "terjadi error pada server"

		if fiberErr, ok := err.(*fiber.Error); ok {
			status = fiberErr.Code
			message = fiberErr.Message
		}

		logger.Error(
			"unhandled_error",
			slog.String("path", c.Path()),
			slog.Int("status", status),
			slog.String("error", err.Error()),
		)

		return helper.Fail(c, status, message)
	}
}
