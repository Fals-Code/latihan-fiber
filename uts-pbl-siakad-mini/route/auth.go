package route

import (
	"github.com/gofiber/fiber/v2"

	"tugas1-go/uts-pbl-siakad-mini/app/handler"
	"tugas1-go/uts-pbl-siakad-mini/middleware"
)

func Register(app *fiber.App, authHandler *handler.AuthHandler, studentHandler *handler.StudentHandler, courseHandler *handler.CourseHandler, secret []byte, accounts middleware.AccountLookup) {
	api := app.Group("/api/v1")
	api.Get("/courses", middleware.RequireAuth(secret, accounts), courseHandler.List)
	auth := api.Group("/auth")
	auth.Post("/login", middleware.LoginFailureLimiter(), authHandler.Login)
	auth.Get("/me", middleware.RequireAuth(secret, accounts), authHandler.Me)
	students := api.Group("/students", middleware.RequireAuth(secret, accounts))
	students.Get("", middleware.RequireAdmin(), studentHandler.List)
	students.Post("", middleware.RequireAdmin(), studentHandler.Create)
	students.Get("/:id", studentHandler.Detail)
	students.Put("/:id", middleware.RequireAdmin(), studentHandler.Update)
	students.Delete("/:id", middleware.RequireAdmin(), studentHandler.Delete)
}
