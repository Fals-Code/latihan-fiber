package handler

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"

	"tugas1-go/uts-pbl-siakad-mini/app/dto"
	"tugas1-go/uts-pbl-siakad-mini/app/model"
	"tugas1-go/uts-pbl-siakad-mini/app/service"
	"tugas1-go/uts-pbl-siakad-mini/middleware"
)

type AuthHandler struct {
	service  *service.AuthService
	accounts interface {
		FindActiveAccount(ctx context.Context, userID int64) (model.CurrentUser, error)
	}
}

func NewAuthHandler(authService *service.AuthService, accounts interface {
	FindActiveAccount(ctx context.Context, userID int64) (model.CurrentUser, error)
}) *AuthHandler {
	return &AuthHandler{service: authService, accounts: accounts}
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var request dto.LoginRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(dto.ErrorResponse{Success: false, Message: "Validasi gagal", Errors: map[string][]string{"body": {"Body harus berupa JSON yang valid"}}})
	}
	if errorsByField := service.ValidateLogin(request.Email, request.Password); errorsByField != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(dto.ErrorResponse{Success: false, Message: "Validasi gagal", Errors: errorsByField})
	}
	result, err := h.service.Login(c.UserContext(), request.Email, request.Password)
	if errors.Is(err, service.ErrInvalidCredentials) {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Success: false, Message: "Email atau password tidak valid"})
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Success: false, Message: "Terjadi kesalahan pada server"})
	}
	return c.Status(fiber.StatusOK).JSON(dto.Envelope{
		Success: true,
		Message: "Login berhasil",
		Data: dto.LoginResponse{
			AccessToken: result.Token,
			TokenType:   "Bearer",
			ExpiresIn:   int64(result.ExpiresIn.Seconds()),
			User:        dto.UserResponse{ID: result.User.ID, Email: result.User.Email, Role: result.User.Role},
		},
	})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	user, ok := c.Locals(middleware.AuthUserLocal).(model.CurrentUser)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Success: false, Message: "Token tidak valid atau akun tidak aktif"})
	}
	response := dto.MeResponse{ID: user.ID, Email: user.Email, Role: user.Role}
	if user.Student != nil {
		response.Student = &dto.StudentDetail{NIM: user.Student.NIM, Nama: user.Student.Nama, Prodi: user.Student.Prodi, Angkatan: user.Student.Angkatan}
	}
	return c.Status(fiber.StatusOK).JSON(dto.Envelope{Success: true, Message: "Data user berhasil diambil", Data: response})
}
