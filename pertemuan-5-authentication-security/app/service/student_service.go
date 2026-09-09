package service

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"tugas1-go/pertemuan-5-authentication-security/app/model"
	"tugas1-go/pertemuan-5-authentication-security/app/repository"
	"tugas1-go/pertemuan-5-authentication-security/helper"
)

type StudentService struct {
	repo repository.StudentRepository
}

func NewStudentService(repo repository.StudentRepository) *StudentService {
	return &StudentService{
		repo: repo,
	}
}

// translateRepositoryError menerjemahkan error aplikasi
// menjadi status HTTP.
func translateRepositoryError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return helper.Fail(
			c,
			fiber.StatusNotFound,
			"mahasiswa tidak ditemukan",
		)

	case errors.Is(err, repository.ErrDuplicate):
		return helper.Fail(
			c,
			fiber.StatusConflict,
			"NIM sudah digunakan",
		)

	default:
		return helper.Fail(
			c,
			fiber.StatusInternalServerError,
			"terjadi kesalahan pada database",
		)
	}
}

// List mengambil daftar mahasiswa.
func (s *StudentService) List(c *fiber.Ctx) error {
	q := helper.ParseListQuery(c)

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	students, total, err := s.repo.FindAll(ctx, q)
	if err != nil {
		return translateRepositoryError(c, err)
	}

	return helper.SuccessList(
		c,
		"daftar mahasiswa berhasil diambil",
		students,
		&model.Meta{
			Page:       q.Page,
			Limit:      q.Limit,
			Total:      total,
			TotalPages: CountTotalPages(total, q.Limit),
		},
	)
}

// Get mengambil satu mahasiswa berdasarkan ID.
func (s *StudentService) Get(c *fiber.Ctx) error {
	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
		)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

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
	var req model.CreateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
		)
	}

	var body map[string]json.RawMessage
	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
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
		return helper.FailValidation(c, errs)
	}

	student := model.Student{
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: true,
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
	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
		)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	// Urutan ini dipertahankan dari Pertemuan 3
	// agar perilaku HTTP tidak berubah.
	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateRepositoryError(c, err)
	}

	var req model.ReplaceStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
		)
	}

	var body map[string]json.RawMessage
	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
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
		return helper.FailValidation(c, errs)
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
	id, valid := helper.ParamID(c)
	if !valid {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"id harus berupa angka positif",
		)
	}

	ctx, cancel := helper.RequestContext(c)
	defer cancel()

	student, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return translateRepositoryError(c, err)
	}

	var req model.PatchStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"body harus berupa JSON yang valid",
		)
	}

	if IsEmptyPatch(req) {
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
			"tidak ada field yang diubah",
		)
	}

	updated, errs := ApplyPatch(student, req)
	if len(errs) > 0 {
		return helper.FailValidation(c, errs)
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
		return helper.Fail(
			c,
			fiber.StatusBadRequest,
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
