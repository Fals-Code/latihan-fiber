package main

import (
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"

	"tugas1-go/uts-pbl-siakad-mini/config"
	"tugas1-go/uts-pbl-siakad-mini/database"
)

func main() {
	config.LoadEnv()

	pool, err := database.NewPool(context.Background())
	if err != nil {
		log.Printf("database initialization failed: %v", err)
		os.Exit(1)
	}
	defer pool.Close()

	app := fiber.New(fiber.Config{AppName: "SIAKAD Mini"})
	if err := app.Listen(":" + config.GetEnv("APP_PORT", "3000")); err != nil {
		log.Printf("server stopped: %v", err)
		os.Exit(1)
	}
}
