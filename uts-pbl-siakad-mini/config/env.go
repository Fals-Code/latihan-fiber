package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

const envFile = "uts-pbl-siakad-mini/.env"

// LoadEnv loads the UTS-specific .env without overriding process environment.
func LoadEnv() {
	if err := godotenv.Load(envFile); err != nil && !os.IsNotExist(err) {
		log.Printf("warning: could not load %s: %v", envFile, err)
	}
}

// GetEnv returns the environment value or fallback when unset or empty.
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
