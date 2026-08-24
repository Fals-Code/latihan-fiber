package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// LoadEnv memuat konfigurasi dari file .env.
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("peringatan: .env tidak ditemukan, memakai environment sistem")
	}
}

// GetEnv mengambil nilai environment atau nilai bawaan.
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}

	return fallback
}

// GetEnvInt mengambil environment dalam bentuk integer.
func GetEnvInt(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("peringatan: %s bukan angka, memakai bawaan %d", key, fallback)
		return fallback
	}

	return parsed
}
