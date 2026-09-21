package service

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"tugas1-go/pertemuan-7-advanced-api-design/app/model"
	"tugas1-go/pertemuan-7-advanced-api-design/app/repository"
	"tugas1-go/pertemuan-7-advanced-api-design/helper"
)

type UserService struct {
	repo  repository.UserRepository
	perms *helper.PermissionSet
}

func NewUserService(repo repository.UserRepository, perms *helper.PermissionSet) *UserService {
	return &UserService{repo: repo, perms: perms}
}

func (s *UserService) current(c *fiber.Ctx) (model.AuthUser, bool) {
	return helper.CurrentUser(c)
}

func userID(c *fiber.Ctx) (int, bool) {
	return helper.ParamID(c)
}

func (s *UserService) List(c *fiber.Ctx) error {
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	users, err := s.repo.FindAll(ctx)
	if err != nil {
		return translateUserRepositoryError(c, err)
	}
	return helper.Success(c, fiber.StatusOK, "daftar pengguna berhasil diambil", users)
}

func (s *UserService) Get(c *fiber.Ctx) error {
	current, ok := s.current(c)
	if !ok {
		return helper.NewAppError(fiber.StatusUnauthorized, "belum terautentikasi")
	}
	id, valid := userID(c)
	if !valid {
		return helper.NewAppError(fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	if !CanAccessUser(current, id, s.perms, "user:read:any") {
		return helper.NewAppError(fiber.StatusForbidden, "tidak memiliki akses ke pengguna ini")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateUserRepositoryError(c, err)
	}
	return helper.Success(c, fiber.StatusOK, "profil pengguna ditemukan", user)
}

func (s *UserService) Create(c *fiber.Ctx) error {
	var req model.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.NewAppError(fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return helper.NewAppError(fiber.StatusUnprocessableEntity, "username, email, dan password wajib diisi")
	}
	role := strings.TrimSpace(req.Role)
	if role == "" {
		role = "user"
	}
	if !s.perms.IsKnownRole(role) {
		return helper.NewAppError(fiber.StatusUnprocessableEntity, "role tidak dikenal")
	}
	hash, err := helper.HashPassword(req.Password)
	if err != nil {
		return helper.NewAppError(fiber.StatusInternalServerError, "gagal memproses password")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	user, err := s.repo.Create(ctx, model.User{Username: req.Username, Email: req.Email, PasswordHash: hash, Role: role, IsActive: true})
	if err != nil {
		return translateUserRepositoryError(c, err)
	}
	return helper.Created(c, "pengguna berhasil dibuat", user, "/api/v1/users/"+strconv.Itoa(user.ID))
}

func (s *UserService) update(c *fiber.Ctx) error {
	current, ok := s.current(c)
	if !ok {
		return helper.NewAppError(fiber.StatusUnauthorized, "belum terautentikasi")
	}
	id, valid := userID(c)
	if !valid {
		return helper.NewAppError(fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	if !CanAccessUser(current, id, s.perms, "user:update:any") {
		return helper.NewAppError(fiber.StatusForbidden, "tidak memiliki akses untuk mengubah pengguna ini")
	}
	var req model.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.NewAppError(fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateUserRepositoryError(c, err)
	}
	if req.Username != nil {
		user.Username = strings.TrimSpace(*req.Username)
	}
	if req.Email != nil {
		user.Email = strings.TrimSpace(*req.Email)
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	user, err = s.repo.Update(ctx, user)
	if err != nil {
		return translateUserRepositoryError(c, err)
	}
	return helper.Success(c, fiber.StatusOK, "pengguna berhasil diperbarui", user)
}

func (s *UserService) Replace(c *fiber.Ctx) error { return s.update(c) }
func (s *UserService) Patch(c *fiber.Ctx) error   { return s.update(c) }

func (s *UserService) AssignRole(c *fiber.Ctx) error {
	current, ok := s.current(c)
	if !ok {
		return helper.NewAppError(fiber.StatusUnauthorized, "belum terautentikasi")
	}
	id, valid := userID(c)
	if !valid {
		return helper.NewAppError(fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	var req model.AssignRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return helper.NewAppError(fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}
	if errs := ValidateAssignRole(current, id, req, s.perms); len(errs) > 0 {
		return helper.NewValidationError(errs)
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	user, err := s.repo.UpdateRole(ctx, id, strings.TrimSpace(req.Role))
	if err != nil {
		return translateUserRepositoryError(c, err)
	}
	return helper.Success(c, fiber.StatusOK, "role pengguna berhasil diperbarui", user)
}

func (s *UserService) Delete(c *fiber.Ctx) error {
	current, ok := s.current(c)
	if !ok {
		return helper.NewAppError(fiber.StatusUnauthorized, "belum terautentikasi")
	}
	id, valid := userID(c)
	if !valid {
		return helper.NewAppError(fiber.StatusBadRequest, "id harus berupa angka positif")
	}
	if current.UserID == id {
		return helper.NewAppError(fiber.StatusForbidden, "tidak boleh menghapus akun sendiri")
	}
	ctx, cancel := helper.RequestContext(c)
	defer cancel()
	if err := s.repo.Delete(ctx, id); err != nil {
		return translateUserRepositoryError(c, err)
	}
	return helper.NoContent(c)
}

func translateUserRepositoryError(c *fiber.Ctx, err error) error {
	if errors.Is(err, repository.ErrNotFound) {
		return helper.NewAppError(fiber.StatusNotFound, "pengguna tidak ditemukan")
	}
	if errors.Is(err, repository.ErrDuplicate) {
		return helper.NewAppError(fiber.StatusConflict, "username atau email sudah digunakan")
	}
	return helper.NewAppError(fiber.StatusInternalServerError, "terjadi kesalahan pada database")
}
