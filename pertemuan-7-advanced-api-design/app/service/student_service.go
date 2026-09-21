package service

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"tugas1-go/pertemuan-7-advanced-api-design/app/model"
	"tugas1-go/pertemuan-7-advanced-api-design/app/repository"
	"tugas1-go/pertemuan-7-advanced-api-design/helper"
)

type StudentService struct {
	repo  repository.StudentRepository
	perms *helper.PermissionSet
}

func NewStudentService(repo repository.StudentRepository, perms *helper.PermissionSet) *StudentService {
	return &StudentService{repo: repo, perms: perms}
}

// translateRepositoryError menerjemahkan error aplikasi
// menjadi status HTTP.
func translateRepositoryError(_ *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.NewAppError(fiber.StatusNotFound,
			"mahasiswa tidak ditemukan",
		)

	case errors.Is(err, repository.ErrDuplicate):
		return helper.NewAppError(fiber.StatusConflict,
			"NIM sudah digunakan",
		)

	default:
		return helper.NewAppError(fiber.StatusInternalServerError,
			"terjadi kesalahan pada database",
		)
	}
}

// List mengambil daftar mahasiswa.
func (s *StudentService) List(c *fiber.Ctx) error {
	q := helper.ParseListQuery(c)

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	var students []model.Student
	var total int
	var err error
	meta := &model.Meta{Page: q.Page, Limit: q.Limit}
	if q.Cursor != "" {
		createdAt, id, decodeErr := helper.DecodeCursor(q.Cursor)
		if decodeErr != nil {
			return helper.NewAppError(fiber.StatusBadRequest, "cursor tidak valid")
		}
		students, err = s.repo.FindAfterCursor(ctx, q, createdAt, id)
	} else {
		students, total, err = s.repo.FindAll(ctx, q)
		meta.Total, meta.TotalPages = total, CountTotalPages(total, q.Limit)
	}
	if err != nil {
		return translateRepositoryError(c, err)
	}
	if len(students) == q.Limit {
		last := students[len(students)-1]
		cursor, encodeErr := helper.EncodeCursor(last.CreatedAt, last.ID)
		if encodeErr != nil {
			return helper.NewAppError(fiber.StatusInternalServerError, "gagal membuat cursor")
		}
		meta.NextCursor = cursor
	}
	return helper.Negotiate(c, fiber.StatusOK, "daftar mahasiswa berhasil diambil", students, meta)
}

// Get mengambil satu mahasiswa berdasarkan ID.
func (s *StudentService) Get(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.NewAppError(fiber.StatusUnauthorized, "belum terautentikasi")
	}
	id, valid := helper.ParamID(c)
	if !valid {
		return helper.NewAppError(fiber.StatusBadRequest,
			"id harus berupa angka positif",
		)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	ownerID, err := s.repo.FindOwnerIDByID(ctx, id)
	if err != nil {
		return translateRepositoryError(c, err)
	}
	if !canAccessStudentOwner(current, ownerID, s.perms, "student:read:any") {
		return helper.NewAppError(fiber.StatusForbidden, "tidak memiliki akses ke mahasiswa ini")
	}
	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateRepositoryError(c, err)
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"mahasiswa ditemukan",
		student,
	)
}

// Create menambahkan mahasiswa baru.
func (s *StudentService) Create(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.NewAppError(fiber.StatusUnauthorized, "belum terautentikasi")
	}
	var req model.CreateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.NewAppError(fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
		)
	}
	if errs := helper.ValidateRequest(req); len(errs) > 0 {
		return helper.NewValidationError(errs)
	}

	var body map[string]json.RawMessage
	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return helper.NewAppError(fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
		)
	}

	fields := map[string]bool{
		"nim":   body["nim"] != nil,
		"name":  body["name"] != nil,
		"grade": body["grade"] != nil,
	}

	req, errs := ValidateCreate(req, fields)
	if len(errs) > 0 {
		return helper.NewValidationError(errs)
	}

	student := model.Student{
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: true,
		OwnerID:  &current.UserID,
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	student, err := s.repo.Create(ctx, student)
	if err != nil {
		return translateRepositoryError(c, err)
	}

	return helper.Created(
		c,
		"mahasiswa berhasil ditambahkan",
		student,
		"/api/v1/students/"+strconv.Itoa(student.ID),
	)
}

// Replace mengganti seluruh data mahasiswa melalui PUT.
func (s *StudentService) Replace(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.NewAppError(fiber.StatusUnauthorized, "belum terautentikasi")
	}
	id, valid := helper.ParamID(c)
	if !valid {
		return helper.NewAppError(fiber.StatusBadRequest,
			"id harus berupa angka positif",
		)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	ownerID, err := s.repo.FindOwnerIDByID(ctx, id)
	if err != nil {
		return translateRepositoryError(c, err)
	}
	if !canAccessStudentOwner(current, ownerID, s.perms, "student:update:any") {
		return helper.NewAppError(fiber.StatusForbidden, "tidak memiliki akses untuk mengubah mahasiswa ini")
	}
	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateRepositoryError(c, err)
	}

	var req model.ReplaceStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.NewAppError(fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
		)
	}
	if errs := helper.ValidateRequest(req); len(errs) > 0 {
		return helper.NewValidationError(errs)
	}

	var body map[string]json.RawMessage
	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return helper.NewAppError(fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
		)
	}

	fields := map[string]bool{
		"nim":       body["nim"] != nil,
		"name":      body["name"] != nil,
		"grade":     body["grade"] != nil,
		"is_active": body["is_active"] != nil,
	}

	req, errs := ValidateReplace(req, fields)
	if len(errs) > 0 {
		return helper.NewValidationError(errs)
	}

	student.NIM = req.NIM
	student.Name = req.Name
	student.Grade = req.Grade
	student.IsActive = req.IsActive

	student, err = s.repo.Update(ctx, student)
	if err != nil {
		return translateRepositoryError(c, err)
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"data mahasiswa berhasil diganti",
		student,
	)
}

// Patch mengubah sebagian data mahasiswa.
func (s *StudentService) Patch(c *fiber.Ctx) error {
	current, ok := helper.CurrentUser(c)
	if !ok {
		return helper.NewAppError(fiber.StatusUnauthorized, "belum terautentikasi")
	}
	id, valid := helper.ParamID(c)
	if !valid {
		return helper.NewAppError(fiber.StatusBadRequest,
			"id harus berupa angka positif",
		)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	ownerID, err := s.repo.FindOwnerIDByID(ctx, id)
	if err != nil {
		return translateRepositoryError(c, err)
	}
	if !canAccessStudentOwner(current, ownerID, s.perms, "student:update:any") {
		return helper.NewAppError(fiber.StatusForbidden, "tidak memiliki akses untuk mengubah mahasiswa ini")
	}
	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateRepositoryError(c, err)
	}

	var req model.PatchStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.NewAppError(fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
		)
	}
	if errs := helper.ValidateRequest(req); len(errs) > 0 {
		return helper.NewValidationError(errs)
	}

	if IsEmptyPatch(req) {
		return helper.NewAppError(fiber.StatusBadRequest,
			"tidak ada field yang diubah",
		)
	}

	updated, errs := ApplyPatch(student, req)
	if len(errs) > 0 {
		return helper.NewValidationError(errs)
	}

	student, err = s.repo.Update(ctx, updated)
	if err != nil {
		return translateRepositoryError(c, err)
	}

	return helper.Success(
		c,
		fiber.StatusOK,
		"data mahasiswa berhasil diperbarui",
		student,
	)
}

// Delete menghapus mahasiswa.
func (s *StudentService) Delete(c *fiber.Ctx) error {
	id, valid := helper.ParamID(c)
	if !valid {
		return helper.NewAppError(fiber.StatusBadRequest,
			"id harus berupa angka positif",
		)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	if err := s.repo.Delete(ctx, id); err != nil {
		return translateRepositoryError(c, err)
	}

	return helper.NoContent(c)
}
