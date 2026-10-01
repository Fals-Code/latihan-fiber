package route

import (
	"github.com/gofiber/fiber/v2"

	"tugas1-go/uts-pbl-siakad-mini/app/handler"
	"tugas1-go/uts-pbl-siakad-mini/middleware"
)

func Register(app *fiber.App, authHandler *handler.AuthHandler, secret []byte, accounts middleware.AccountLookup) {
	api := app.Group("/api/v1")
	auth := api.Group("/auth")
	auth.Post("/login", middleware.LoginFailureLimiter(), authHandler.Login)
	auth.Get("/me", middleware.RequireAuth(secret, accounts), authHandler.Me)
}
