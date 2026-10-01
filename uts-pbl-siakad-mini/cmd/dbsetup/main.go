package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"tugas1-go/uts-pbl-siakad-mini/config"
	"tugas1-go/uts-pbl-siakad-mini/database"
	"tugas1-go/uts-pbl-siakad-mini/seeders"
)

func main() {
	if len(os.Args) != 2 || (os.Args[1] != "migrate" && os.Args[1] != "seed") {
		log.Fatal("usage: go run ./uts-pbl-siakad-mini/cmd/dbsetup [migrate|seed]")
	}
	config.LoadEnv()
	ctx := context.Background()
	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	switch os.Args[1] {
	case "migrate":
		migrationDir := filepath.Join("uts-pbl-siakad-mini", "migrations")
		if err := database.Migrate(ctx, pool, migrationDir); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Migrations applied.")
	case "seed":
		if err := seeders.Seed(ctx, pool); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Seed data inserted.")
	}
}
