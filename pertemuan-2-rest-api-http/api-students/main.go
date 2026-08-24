package main

import (
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
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
		return ok(c, "server berjalan", fiber.Map{
			"timestamp": time.Now(),
		})
	})

	studentsAPI := api.Group("/students", requireJSON)

	studentsAPI.Get("/", listStudents)
	studentsAPI.Get("/:id", getStudent)
	studentsAPI.Post("/", createStudent)
	studentsAPI.Put("/:id", replaceStudent)
	studentsAPI.Patch("/:id", patchStudent)
	studentsAPI.Delete("/:id", deleteStudent)

	// Endpoint yang tidak tersedia
	app.Use(func(c *fiber.Ctx) error {
		return fail(c, fiber.StatusNotFound, "endpoint tidak ditemukan")
	})

	log.Println("Server berjalan di http://localhost:3000")
	log.Fatal(app.Listen(":3000"))
}
