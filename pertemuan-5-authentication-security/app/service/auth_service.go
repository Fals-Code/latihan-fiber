package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"tugas1-go/pertemuan-5-authentication-security/app/model"
	"tugas1-go/pertemuan-5-authentication-security/app/repository"
	"tugas1-go/pertemuan-5-authentication-security/helper"
)

const invalidCredentialsMessage = "username atau password salah"

var ErrAuthenticatedUserMissing = errors.New("authenticated user missing")

type AuthService struct {
	users      repository.UserRepository
	tokens     repository.TokenRepository
	jwt        *helper.JWTManager
	refreshTTL time.Duration
}

func NewAuthService(users repository.UserRepository, tokens repository.TokenRepository, jwtManager *helper.JWTManager, refreshTTL time.Duration) *AuthService {
	return &AuthService{users: users, tokens: tokens, jwt: jwtManager, refreshTTL: refreshTTL}
}

func (s *AuthService) Register(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	if errs := ValidateRegister(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}
	passwordHash, err := helper.HashPassword(req.Password)
	if err != nil {
		return helper.Fail(c, fiber.StatusInternalServerError, "gagal memproses password")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	user, err := s.users.Create(ctx, model.User{Username: req.Username, Email: req.Email, PasswordHash: passwordHash, Role: "user", IsActive: true})
	if err != nil {
		return translateAuthDatabaseError(c, err)
	}
	return helper.Created(c, "registrasi berhasil", user, "/api/v1/auth/me")
}

func (s *AuthService) Login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	if errs := ValidateLogin(req); len(errs) > 0 {
		return helper.FailValidation(c, errs)
	}
	req.Username = strings.TrimSpace(req.Username)
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	user, err := s.users.FindByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			helper.VerifyDummyPassword(req.Password)
			return helper.Fail(c, fiber.StatusUnauthorized, invalidCredentialsMessage)
		}
		return translateAuthDatabaseError(c, err)
	}
	if !helper.VerifyPassword(user.PasswordHash, req.Password) {
		return helper.Fail(c, fiber.StatusUnauthorized, invalidCredentialsMessage)
	}
	if !user.IsActive {
		return helper.Fail(c, fiber.StatusForbidden, "akun dinonaktifkan")
	}
	pair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return translateAuthDatabaseError(c, err)
	}
	return helper.Success(c, fiber.StatusOK, "login berhasil", pair)
}

func (s *AuthService) issueTokenPair(ctx context.Context, user model.User) (model.TokenPair, error) {
	accessToken, err := s.jwt.GenerateAccessToken(user)
	if err != nil {
		return model.TokenPair{}, err
	}
	rawRefreshToken, err := helper.RandomToken(32)
	if err != nil {
		return model.TokenPair{}, err
	}
	expiresAt := time.Now().Add(s.refreshTTL)
	if err := s.tokens.Save(ctx, model.RefreshToken{UserID: user.ID, TokenHash: helper.SHA256Hex(rawRefreshToken), ExpiresAt: expiresAt}); err != nil {
		return model.TokenPair{}, err
	}
	return model.TokenPair{AccessToken: accessToken, RefreshToken: rawRefreshToken, TokenType: "Bearer", ExpiresIn: int(s.jwt.AccessTTL().Seconds())}, nil
}

func (s *AuthService) Refresh(c *fiber.Ctx) error {
	var req model.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	if strings.TrimSpace(req.RefreshToken) == "" {
		return helper.Fail(c, fiber.StatusUnauthorized, "refresh token tidak valid atau sudah kedaluwarsa")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	stored, err := s.tokens.FindActive(ctx, helper.SHA256Hex(req.RefreshToken))
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return helper.Fail(c, fiber.StatusUnauthorized, "refresh token tidak valid atau sudah kedaluwarsa")
		}
		return translateAuthDatabaseError(c, err)
	}
	user, err := s.users.FindByID(ctx, stored.UserID)
	if err != nil || !user.IsActive {
		if err != nil && !errors.Is(err, repository.ErrNotFound) {
			return translateAuthDatabaseError(c, err)
		}
		return helper.Fail(c, fiber.StatusUnauthorized, "refresh token tidak valid atau sudah kedaluwarsa")
	}
	if err := s.tokens.Revoke(ctx, stored.TokenHash); err != nil {
		return translateAuthDatabaseError(c, err)
	}
	pair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return translateAuthDatabaseError(c, err)
	}
	return helper.Success(c, fiber.StatusOK, "token berhasil diperbarui", pair)
}

func (s *AuthService) Logout(c *fiber.Ctx) error {
	var req model.RefreshRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	if strings.TrimSpace(req.RefreshToken) != "" {
		ctx, cancel := helper.RequestContext(c)
		defer cancel()
		if err := s.tokens.Revoke(ctx, helper.SHA256Hex(req.RefreshToken)); err != nil {
			return translateAuthDatabaseError(c, err)
		}
	}
	return helper.Success(c, fiber.StatusOK, "logout berhasil", nil)
}

func (s *AuthService) Me(c *fiber.Ctx) error {
	identity, ok := helper.CurrentUser(c)
	if !ok {
		return helper.Fail(c, fiber.StatusUnauthorized, "pengguna belum terautentikasi")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	user, err := s.users.FindByID(ctx, identity.UserID)
	if err != nil {
		return translateAuthDatabaseError(c, err)
	}
	if !user.IsActive {
		return helper.Fail(c, fiber.StatusUnauthorized, "akun dinonaktifkan")
	}
	return helper.Success(c, fiber.StatusOK, "profil pengguna ditemukan", user)
}

func translateAuthDatabaseError(c *fiber.Ctx, err error) error {
	if errors.Is(err, repository.ErrDuplicate) {
		return helper.Fail(c, fiber.StatusConflict, "username atau email sudah digunakan")
	}
	return helper.Fail(c, fiber.StatusInternalServerError, "terjadi kesalahan pada database")
}
