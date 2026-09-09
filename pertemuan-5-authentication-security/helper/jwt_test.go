package helper

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"tugas1-go/pertemuan-5-authentication-security/app/model"
)

func TestJWTManagerGeneratesAndParsesToken(t *testing.T) {
	manager := NewJWTManager("12345678901234567890123456789012", "test-issuer", time.Hour)
	token, err := manager.GenerateAccessToken(model.User{ID: 7, Username: "falah", Role: "user"})
	if err != nil {
		t.Fatal(err)
	}
	identity, err := manager.ParseAccessToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if identity.UserID != 7 || identity.Username != "falah" || identity.Role != "user" {
		t.Fatalf("unexpected identity: %+v", identity)
	}
}

func TestJWTManagerRejectsWrongSecretAndUnexpectedAlgorithm(t *testing.T) {
	manager := NewJWTManager("12345678901234567890123456789012", "test-issuer", time.Hour)
	token, err := manager.GenerateAccessToken(model.User{ID: 7, Username: "falah", Role: "user"})
	if err != nil {
		t.Fatal(err)
	}
	wrong := NewJWTManager("abcdefghijklmnopqrstuvwxyz123456", "test-issuer", time.Hour)
	if _, err := wrong.ParseAccessToken(token); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected invalid token, got %v", err)
	}
	claims := jwt.MapClaims{"sub": "7", "iss": "test-issuer", "username": "falah", "role": "user", "iat": time.Now().Unix(), "exp": time.Now().Add(time.Hour).Unix()}
	none, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.ParseAccessToken(none); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected algorithm rejection, got %v", err)
	}
}

func TestJWTManagerMapsExpiredAndMalformedTokens(t *testing.T) {
	manager := NewJWTManager("12345678901234567890123456789012", "test-issuer", -time.Second)
	expired, err := manager.GenerateAccessToken(model.User{ID: 7, Username: "falah", Role: "user"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.ParseAccessToken(expired); !errors.Is(err, ErrExpiredToken) {
		t.Fatalf("expected expired token, got %v", err)
	}
	if _, err := manager.ParseAccessToken("not-a-token"); !errors.Is(err, ErrInvalidToken) {
		t.Fatalf("expected invalid token, got %v", err)
	}
}
