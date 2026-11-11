package route

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/gofiber/swagger"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"tugas1-go/pertemuan-12-testing-swagger-docker/app/service"
	"tugas1-go/pertemuan-12-testing-swagger-docker/helper"
	"tugas1-go/pertemuan-12-testing-swagger-docker/middleware"
)

type Dependencies struct {
	Mongo               *mongo.Client
	CourseService       *service.CourseService
	ReviewService       *service.ReviewService
	NotificationService *service.NotificationService
	Pool                *pgxpool.Pool
	JWT                 *helper.JWTManager
	Permissions         *helper.PermissionSet
	StudentService      *service.StudentService
	AchievementService  *service.AchievementService
	AuthService         *service.AuthService
	UserService         *service.UserService
	EnableSwagger       bool
}

func Register(app *fiber.App, deps Dependencies) {
	if deps.EnableSwagger {
		app.Get("/swagger/*", fiberSwagger.HandlerDefault)
	}
	api := app.Group("/api/v1")
	api.Get("/health", healthCheck(deps.Pool, deps.Mongo))

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
	courses := api.Group("/courses", middleware.RequireJSON, requireAuth)
	courses.Get("/", deps.CourseService.List)
	// Literal /mine precedes /:id so it cannot be parsed as an integer ID.
	courses.Get("/mine", deps.CourseService.Mine)
	courses.Get("/:id", deps.CourseService.Get)
	courses.Post("/:id/enroll", deps.CourseService.Enroll)
	courses.Delete("/:id/enroll", deps.CourseService.Drop)
	courses.Patch("/:id", middleware.RequirePermission(deps.Permissions, "course:manage"), deps.CourseService.UpdateQuota)
	courses.Get("/:id/reviews", deps.ReviewService.ListByCourse)
	courses.Get("/:id/reviews/summary", deps.ReviewService.Summary)
	courses.Post("/:id/reviews", deps.ReviewService.Create)
	enrollments := api.Group("/enrollments", middleware.RequireJSON, requireAuth)
	enrollments.Post("/swap", deps.CourseService.Swap)
	reviews := api.Group("/reviews", middleware.RequireJSON, requireAuth)
	reviews.Get("/:id", deps.ReviewService.Get)
	reviews.Patch("/:id", deps.ReviewService.Update)
	reviews.Delete("/:id", deps.ReviewService.Delete)
	reviews.Post("/:id/replies", deps.ReviewService.Reply)
	reviews.Post("/:id/helpful", deps.ReviewService.MarkHelpful)
	notifications := api.Group("/notifications", middleware.RequireJSON, requireAuth)
	notifications.Get("/mine", deps.NotificationService.Mine)
	admin := api.Group("/admin", middleware.RequireJSON, requireAuth)
	admin.Get("/queue", middleware.RequirePermission(deps.Permissions, "queue:monitor"), deps.NotificationService.Monitor)
}

func healthCheck(pool *pgxpool.Pool, mongoClient *mongo.Client) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			return helper.ServiceUnavailable("PostgreSQL tidak dapat dihubungi")
		}
		if err := mongoClient.Ping(ctx, nil); err != nil {
			return helper.ServiceUnavailable("MongoDB tidak dapat dihubungi")
		}
		return helper.Success(c, fiber.StatusOK, "server, PostgreSQL, dan MongoDB berjalan", nil)
	}
}
