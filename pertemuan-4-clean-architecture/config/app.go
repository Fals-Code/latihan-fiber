package config

import (
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"tugas1-go/pertemuan-4-clean-architecture/app/service"
	"tugas1-go/pertemuan-4-clean-architecture/helper"
	"tugas1-go/pertemuan-4-clean-architecture/middleware"
	"tugas1-go/pertemuan-4-clean-architecture/route"
)

// NewApp merakit aplikasi Fiber, middleware, dan route.
func NewApp(
	logger *slog.Logger,
	pool *pgxpool.Pool,
	studentService *service.StudentService,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      GetEnv("APP_NAME", "Praktikum Backend Lanjut"),
		ErrorHandler: newErrorHandler(logger),
	})

	middleware.Register(app, logger)
	route.Register(app, pool, studentService)

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
