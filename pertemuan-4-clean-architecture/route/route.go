package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"tugas1-go/pertemuan-4-clean-architecture/app/service"
	"tugas1-go/pertemuan-4-clean-architecture/helper"
	"tugas1-go/pertemuan-4-clean-architecture/middleware"
)

// Register mendaftarkan seluruh endpoint aplikasi.
func Register(
	app *fiber.App,
	pool *pgxpool.Pool,
	studentService *service.StudentService,
	achievementService *service.AchievementService,
) {
	api := app.Group("/api/v1")

	api.Get("/health", healthCheck(pool))

	students := api.Group("/students", middleware.RequireJSON)

	students.Get("/", studentService.List)
	students.Get("/:id", studentService.Get)
	students.Post("/", studentService.Create)
	students.Put("/:id", studentService.Replace)
	students.Patch("/:id", studentService.Patch)
	students.Delete("/:id", studentService.Delete)

	achievements := api.Group("/achievements", middleware.RequireJSON)
	achievements.Get("/", achievementService.List)
	achievements.Get("/:id", achievementService.Get)
	achievements.Post("/", achievementService.Create)
	achievements.Put("/:id", achievementService.Replace)
	achievements.Patch("/:id", achievementService.Patch)
	achievements.Delete("/:id", achievementService.Delete)
}

// healthCheck memeriksa apakah server dapat berkomunikasi
// dengan PostgreSQL.
func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(
			c.UserContext(),
			2*time.Second,
		)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			return helper.Fail(
				c,
				fiber.StatusServiceUnavailable,
				"database tidak tersedia",
			)
		}

		return helper.Success(
			c,
			fiber.StatusOK,
			"server dan database berjalan",
			nil,
		)
	}
}
