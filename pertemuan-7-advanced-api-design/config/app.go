package config

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v2"
	"tugas1-go/pertemuan-7-advanced-api-design/app/model"
	"tugas1-go/pertemuan-7-advanced-api-design/helper"
	"tugas1-go/pertemuan-7-advanced-api-design/middleware"
	"tugas1-go/pertemuan-7-advanced-api-design/route"
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
		return helper.NewAppError(fiber.StatusNotFound, "endpoint tidak ditemukan")
	})

	return app
}

// newErrorHandler menjadi jaring pengaman terakhir untuk error
// yang belum ditangani oleh service.
func newErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		status := fiber.StatusInternalServerError
		message := "terjadi error pada server"
		code := helper.CodeInternal
		var fields map[string]string
		var appErr *helper.AppError
		if errors.As(err, &appErr) {
			status, message, code, fields = appErr.Status, appErr.Message, appErr.Code, appErr.Fields
		} else if fiberErr, ok := err.(*fiber.Error); ok {
			status, message, code = fiberErr.Code, fiberErr.Message, helper.CodeBadRequest
		}
		requestID, _ := c.Locals("requestid").(string)
		attrs := []any{slog.String("request_id", requestID), slog.String("path", c.Path()), slog.Int("status", status)}
		if status >= 400 && status < 500 {
			logger.Warn("request_rejected", attrs...)
		} else {
			if appErr != nil && appErr.Cause() != nil {
				attrs = append(attrs, slog.String("error", appErr.Cause().Error()))
			} else {
				attrs = append(attrs, slog.String("error", err.Error()))
			}
			logger.Error("request_failed", attrs...)
		}
		return c.Status(status).JSON(model.FailureResponse{Success: false, Code: code, Message: message, Fields: fields, RequestID: requestID})
	}
}
