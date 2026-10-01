package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

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

func JWTConfig() (string, time.Duration, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", 0, fmt.Errorf("JWT_SECRET is required")
	}
	expiresIn, err := strconv.ParseInt(os.Getenv("JWT_EXPIRES_IN"), 10, 64)
	if err != nil || expiresIn <= 0 {
		return "", 0, fmt.Errorf("JWT_EXPIRES_IN must be a positive number of seconds")
	}
	return secret, time.Duration(expiresIn) * time.Second, nil
}
