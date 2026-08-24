package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"tugas1-go/pertemuan-3-database-repository/app/repository"
	"tugas1-go/pertemuan-3-database-repository/config"
	"tugas1-go/pertemuan-3-database-repository/database"
)

var metodeBerbody = map[string]bool{
	fiber.MethodPost:  true,
	fiber.MethodPut:   true,
	fiber.MethodPatch: true,
}

// Memastikan request body menggunakan JSON
func requireJSON(c *fiber.Ctx) error {
	if metodeBerbody[c.Method()] {
		contentType := c.Get("Content-Type")

		if !strings.HasPrefix(contentType, fiber.MIMEApplicationJSON) {
			return fail(
				c,
				fiber.StatusUnsupportedMediaType,
				"Content-Type harus application/json",
			)
		}
	}

	return c.Next()
}

func main() {
	config.LoadEnv()

	pool, err := database.NewPool(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	studentRepo := repository.NewStudentRepository(pool)
	studentHandler := NewStudentHandler(studentRepo)

	app := fiber.New(fiber.Config{
		AppName: "Student REST API",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			status := fiber.StatusInternalServerError
			message := "terjadi kesalahan pada server"

			if e, ok := err.(*fiber.Error); ok {
				status = e.Code
				message = e.Message
			}

			return fail(c, status, message)
		},
	})

	// Middleware
	app.Use(requestid.New())

	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${locals:requestid} ${method} ${path} ${status} ${latency}\n",
	}))

	app.Use(cors.New())

	api := app.Group("/api/v1")

	api.Get("/health", func(c *fiber.Ctx) error {
		pingCtx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(pingCtx); err != nil {
			return fail(
				c,
				fiber.StatusServiceUnavailable,
				"database tidak tersedia",
			)
		}

		return ok(c, "server berjalan", fiber.Map{
			"timestamp": time.Now(),
		})
	})

	studentsAPI := api.Group("/students", requireJSON)

	studentsAPI.Get("/", studentHandler.List)
	studentsAPI.Get("/:id", studentHandler.Get)
	studentsAPI.Post("/", studentHandler.Create)
	studentsAPI.Put("/:id", studentHandler.Replace)
	studentsAPI.Patch("/:id", studentHandler.Patch)
	studentsAPI.Delete("/:id", studentHandler.Delete)

	// Endpoint yang tidak tersedia
	app.Use(func(c *fiber.Ctx) error {
		return fail(c, fiber.StatusNotFound, "endpoint tidak ditemukan")
	})

	port := config.GetEnv("APP_PORT", "3000")

	log.Printf("Server berjalan di http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}
