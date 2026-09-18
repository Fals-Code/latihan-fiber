package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"tugas1-go/pertemuan-6-authorization-rbac/app/service"
	"tugas1-go/pertemuan-6-authorization-rbac/helper"
	"tugas1-go/pertemuan-6-authorization-rbac/middleware"
)

type Dependencies struct {
	Pool               *pgxpool.Pool
	JWT                *helper.JWTManager
	Permissions        *helper.PermissionSet
	StudentService     *service.StudentService
	AchievementService *service.AchievementService
	AuthService        *service.AuthService
	UserService        *service.UserService
}

func Register(app *fiber.App, deps Dependencies) {
	api := app.Group("/api/v1")
	api.Get("/health", healthCheck(deps.Pool))

	auth := api.Group("/auth", middleware.RequireJSON)
	auth.Post("/register", deps.AuthService.Register)
	auth.Post("/login", middleware.LoginRateLimiter(), deps.AuthService.Login)
	auth.Post("/refresh", deps.AuthService.Refresh)
	auth.Post("/logout", deps.AuthService.Logout)
	auth.Get("/me", middleware.RequireAuth(deps.JWT), deps.AuthService.Me)

	requireAuth := middleware.RequireAuth(deps.JWT)
	users := api.Group("/users", middleware.RequireJSON, requireAuth)
	users.Get("/", middleware.RequirePermission(deps.Permissions, "user:list"), deps.UserService.List)
	users.Post("/", middleware.RequirePermission(deps.Permissions, "user:update:any"), deps.UserService.Create)
	users.Get("/:id", deps.UserService.Get)
	users.Put("/:id", deps.UserService.Replace)
	users.Patch("/:id", deps.UserService.Patch)
	users.Delete("/:id", middleware.RequirePermission(deps.Permissions, "user:delete"), deps.UserService.Delete)
	users.Patch("/:id/role", middleware.RequirePermission(deps.Permissions, "role:assign"), deps.UserService.AssignRole)
	students := api.Group("/students", middleware.RequireJSON, requireAuth)
	students.Get("/", middleware.RequirePermission(deps.Permissions, "student:list"), deps.StudentService.List)
	students.Get("/:id", deps.StudentService.Get)
	students.Post("/", middleware.RequirePermission(deps.Permissions, "student:create"), deps.StudentService.Create)
	students.Put("/:id", deps.StudentService.Replace)
	students.Patch("/:id", deps.StudentService.Patch)
	students.Delete("/:id", middleware.RequirePermission(deps.Permissions, "student:delete"), deps.StudentService.Delete)

	achievements := api.Group("/achievements", middleware.RequireJSON, requireAuth)
	achievements.Get("/", deps.AchievementService.List)
	achievements.Get("/:id", deps.AchievementService.Get)
	achievements.Post("/", deps.AchievementService.Create)
	achievements.Put("/:id", deps.AchievementService.Replace)
	achievements.Patch("/:id", deps.AchievementService.Patch)
	achievements.Delete("/:id", deps.AchievementService.Delete)
}

func healthCheck(pool *pgxpool.Pool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			return helper.Fail(c, fiber.StatusServiceUnavailable, "database tidak tersedia")
		}
		return helper.Success(c, fiber.StatusOK, "server dan database berjalan", nil)
	}
}
