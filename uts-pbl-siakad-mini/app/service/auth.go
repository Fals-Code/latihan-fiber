package service

import (
	"context"
	"errors"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"tugas1-go/uts-pbl-siakad-mini/app/model"
	"tugas1-go/uts-pbl-siakad-mini/app/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidInput       = errors.New("invalid login input")
	ErrInvalidIdentity    = errors.New("invalid identity")
)

type AuthStore interface {
	FindUserByEmail(context.Context, string) (model.User, error)
	FindActiveAccount(context.Context, int64) (model.CurrentUser, error)
}

type AuthService struct {
	store  AuthStore
	secret []byte
	ttl    time.Duration
}

type AuthResult struct {
	Token     string
	Expiry    time.Time
	ExpiresIn time.Duration
	User      model.User
}

func NewAuthService(store AuthStore, secret string, ttl time.Duration) *AuthService {
	return &AuthService{store: store, secret: []byte(secret), ttl: ttl}
}

func ValidateLogin(email, password string) map[string][]string {
	errorsByField := make(map[string][]string)
	if strings.TrimSpace(email) == "" {
		errorsByField["email"] = append(errorsByField["email"], "Email wajib diisi")
	} else if parsed, err := mail.ParseAddress(email); err != nil || parsed.Address != email {
		errorsByField["email"] = append(errorsByField["email"], "Format email tidak valid")
	}
	if password == "" {
		errorsByField["password"] = append(errorsByField["password"], "Password wajib diisi")
	} else if len(password) < 8 {
		errorsByField["password"] = append(errorsByField["password"], "Password minimal 8 karakter")
	}
	if len(errorsByField) == 0 {
		return nil
	}
	return errorsByField
}

func (s *AuthService) Login(ctx context.Context, email, password string) (AuthResult, error) {
	if ValidateLogin(email, password) != nil {
		return AuthResult{}, ErrInvalidInput
	}
	user, err := s.store.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return AuthResult{}, ErrInvalidCredentials
		}
		return AuthResult{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return AuthResult{}, ErrInvalidCredentials
	}
	if user.Role != "admin" && user.Role != "mahasiswa" {
		return AuthResult{}, ErrInvalidCredentials
	}
	if user.Role == "mahasiswa" {
		if _, err := s.store.FindActiveAccount(ctx, user.ID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return AuthResult{}, ErrInvalidCredentials
			}
			return AuthResult{}, err
		}
	}
	if len(s.secret) == 0 || s.ttl <= 0 {
		return AuthResult{}, errors.New("authentication configuration invalid")
	}
	expiresAt := time.Now().Add(s.ttl)
	claims := jwt.RegisteredClaims{
		Subject:   int64String(user.ID),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  claims.Subject,
		"exp":  claims.ExpiresAt.Unix(),
		"iat":  claims.IssuedAt.Unix(),
		"role": user.Role,
	})
	raw, err := token.SignedString(s.secret)
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{Token: raw, Expiry: expiresAt, ExpiresIn: s.ttl, User: user}, nil
}

func ParseToken(raw string, secret []byte) (model.AuthenticatedUser, error) {
	if len(secret) == 0 || strings.TrimSpace(raw) == "" {
		return model.AuthenticatedUser{}, ErrInvalidIdentity
	}
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidIdentity
		}
		return secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		return model.AuthenticatedUser{}, ErrInvalidIdentity
	}
	subject, err := claims.GetSubject()
	if err != nil {
		return model.AuthenticatedUser{}, ErrInvalidIdentity
	}
	id, err := parsePositiveID(subject)
	if err != nil {
		return model.AuthenticatedUser{}, ErrInvalidIdentity
	}
	role, ok := claims["role"].(string)
	if !ok || (role != "admin" && role != "mahasiswa") {
		return model.AuthenticatedUser{}, ErrInvalidIdentity
	}
	return model.AuthenticatedUser{UserID: id, Role: role}, nil
}

func int64String(id int64) string {
	return strconv.FormatInt(id, 10)
}

func parsePositiveID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, ErrInvalidIdentity
	}
	return id, nil
}
