package main

import (
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"

	"tugas1-go/uts-pbl-siakad-mini/app/handler"
	"tugas1-go/uts-pbl-siakad-mini/app/repository"
	"tugas1-go/uts-pbl-siakad-mini/app/service"
	"tugas1-go/uts-pbl-siakad-mini/config"
	"tugas1-go/uts-pbl-siakad-mini/database"
	"tugas1-go/uts-pbl-siakad-mini/route"
)

func main() {
	config.LoadEnv()

	secret, ttl, err := config.JWTConfig()
	if err != nil {
		log.Printf("authentication configuration failed: %v", err)
		os.Exit(1)
	}

	pool, err := database.NewPool(context.Background())
	if err != nil {
		log.Printf("database initialization failed: %v", err)
		os.Exit(1)
	}
	defer pool.Close()

	authRepository := repository.NewAuthRepository(pool)
	authService := service.NewAuthService(authRepository, secret, ttl)
	authHandler := handler.NewAuthHandler(authService, authRepository)
	studentRepository := repository.NewStudentRepository(pool)
	studentService := service.NewStudentService(studentRepository)
	studentHandler := handler.NewStudentHandler(studentService)

	app := fiber.New(fiber.Config{AppName: "SIAKAD Mini"})
	route.Register(app, authHandler, studentHandler, []byte(secret), authRepository)
	if err := app.Listen(":" + config.GetEnv("APP_PORT", "3000")); err != nil {
		log.Printf("server stopped: %v", err)
		os.Exit(1)
	}
}
