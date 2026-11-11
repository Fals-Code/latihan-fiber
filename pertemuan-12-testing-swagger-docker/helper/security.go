package helper

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

const passwordCost = 12

const dummyPasswordHash = "$2a$12$LuB0bcfI20qa79geSYTYPeCX6KrDmWu8GMDWXyyJe7XIsxj4NVwjK"

func HashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), passwordCost)
	return string(hash), err
}

func VerifyPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

func VerifyDummyPassword(plain string) {
	_ = bcrypt.CompareHashAndPassword([]byte(dummyPasswordHash), []byte(plain))
}

func RandomToken(numBytes int) (string, error) {
	bytes := make([]byte, numBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func SHA256Hex(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}
