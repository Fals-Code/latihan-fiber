package helper

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"tugas1-go/pertemuan-5-authentication-security/app/model"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
)

type JWTManager struct {
	secret    []byte
	issuer    string
	accessTTL time.Duration
}

func NewJWTManager(secret, issuer string, accessTTL time.Duration) *JWTManager {
	return &JWTManager{secret: []byte(secret), issuer: issuer, accessTTL: accessTTL}
}

func (m *JWTManager) AccessTTL() time.Duration {
	return m.accessTTL
}

func (m *JWTManager) GenerateAccessToken(user model.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      strconv.Itoa(user.ID),
		"iss":      m.issuer,
		"iat":      now.Unix(),
		"exp":      now.Add(m.accessTTL).Unix(),
		"username": user.Username,
		"role":     user.Role,
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *JWTManager) ParseAccessToken(raw string) (model.AuthUser, error) {
	var claims jwt.MapClaims
	token, err := jwt.ParseWithClaims(raw, &claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	}, jwt.WithIssuer(m.issuer), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return model.AuthUser{}, ErrExpiredToken
		}
		return model.AuthUser{}, ErrInvalidToken
	}
	if !token.Valid {
		return model.AuthUser{}, ErrInvalidToken
	}

	subject, err := claims.GetSubject()
	if err != nil {
		return model.AuthUser{}, ErrInvalidToken
	}
	userID, err := strconv.Atoi(subject)
	if err != nil || userID < 1 {
		return model.AuthUser{}, ErrInvalidToken
	}
	username, ok := claims["username"].(string)
	if !ok || username == "" {
		return model.AuthUser{}, ErrInvalidToken
	}
	role, ok := claims["role"].(string)
	if !ok || role == "" {
		return model.AuthUser{}, ErrInvalidToken
	}

	return model.AuthUser{UserID: userID, Username: username, Role: role}, nil
}
