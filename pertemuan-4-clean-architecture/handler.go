package main

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"tugas1-go/pertemuan-4-clean-architecture/app/model"
	"tugas1-go/pertemuan-4-clean-architecture/app/repository"
)

type StudentHandler struct {
	repo repository.StudentRepository
}

func NewStudentHandler(repo repository.StudentRepository) *StudentHandler {
	return &StudentHandler{
		repo: repo,
	}
}

// Membaca ID dari URL
func paramID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil || id < 1 {
		return 0, false
	}

	return id, true
}

// Mengubah error repository menjadi respons HTTP
func handleRepositoryError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return fail(c, fiber.StatusNotFound, "mahasiswa tidak ditemukan")

	case errors.Is(err, repository.ErrDuplicate):
		return fail(c, fiber.StatusConflict, "NIM sudah digunakan")

	default:
		return fail(c, fiber.StatusInternalServerError, "terjadi kesalahan pada database")
	}
}

// Mengambil daftar mahasiswa
func (h *StudentHandler) List(c *fiber.Ctx) error {
	q := parseListQuery(c)

	ctx, cancel := requestContext(c)
	defer cancel()

	students, total, err := h.repo.FindAll(ctx, q)
	if err != nil {
		return handleRepositoryError(c, err)
	}

	totalPages := (total + q.Limit - 1) / q.Limit

	return okList(
		c,
		"daftar mahasiswa berhasil diambil",
		students,
		&model.Meta{
			Page:       q.Page,
			Limit:      q.Limit,
			Total:      total,
			TotalPages: totalPages,
		},
	)
}

// Mengambil satu mahasiswa berdasarkan ID
func (h *StudentHandler) Get(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	ctx, cancel := requestContext(c)
	defer cancel()

	student, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return handleRepositoryError(c, err)
	}

	return ok(c, "mahasiswa ditemukan", student)
}

// Menambahkan mahasiswa baru
func (h *StudentHandler) Create(c *fiber.Ctx) error {
	var req model.CreateStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	var body map[string]json.RawMessage
	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	errs := map[string]string{}

	if _, ada := body["nim"]; !ada {
		errs["nim"] = "wajib diisi"
	} else if req.NIM <= 0 {
		errs["nim"] = "harus berupa angka positif"
	}

	if _, ada := body["name"]; !ada {
		errs["name"] = "wajib diisi"
	} else {
		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" {
			errs["name"] = "wajib diisi"
		}
	}

	if _, ada := body["grade"]; !ada {
		errs["grade"] = "wajib diisi"
	} else if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "harus berada antara 0 sampai 100"
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	student := model.Student{
		NIM:      req.NIM,
		Name:     req.Name,
		Grade:    req.Grade,
		IsActive: true,
	}

	ctx, cancel := requestContext(c)
	defer cancel()

	student, err := h.repo.Create(ctx, student)
	if err != nil {
		return handleRepositoryError(c, err)
	}

	return created(
		c,
		"mahasiswa berhasil ditambahkan",
		student,
		"/api/v1/students/"+strconv.Itoa(student.ID),
	)
}

// Mengganti seluruh data mahasiswa
func (h *StudentHandler) Replace(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	ctx, cancel := requestContext(c)
	defer cancel()

	student, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return handleRepositoryError(c, err)
	}

	var req model.ReplaceStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	var body map[string]json.RawMessage
	if err := json.Unmarshal(c.Body(), &body); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	errs := map[string]string{}

	if _, ada := body["nim"]; !ada {
		errs["nim"] = "wajib diisi pada PUT"
	} else if req.NIM <= 0 {
		errs["nim"] = "harus berupa angka positif"
	}

	if _, ada := body["name"]; !ada {
		errs["name"] = "wajib diisi pada PUT"
	} else {
		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" {
			errs["name"] = "tidak boleh kosong"
		}
	}

	if _, ada := body["grade"]; !ada {
		errs["grade"] = "wajib diisi pada PUT"
	} else if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "harus berada antara 0 sampai 100"
	}

	if _, ada := body["is_active"]; !ada {
		errs["is_active"] = "wajib diisi pada PUT"
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	student.NIM = req.NIM
	student.Name = req.Name
	student.Grade = req.Grade
	student.IsActive = req.IsActive

	student, err = h.repo.Update(ctx, student)
	if err != nil {
		return handleRepositoryError(c, err)
	}

	return ok(c, "data mahasiswa berhasil diganti", student)
}

// Mengubah sebagian data mahasiswa
func (h *StudentHandler) Patch(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	ctx, cancel := requestContext(c)
	defer cancel()

	student, err := h.repo.FindByID(ctx, id)
	if err != nil {
		return handleRepositoryError(c, err)
	}

	var req model.PatchStudentRequest

	if err := c.BodyParser(&req); err != nil {
		return fail(c, fiber.StatusBadRequest, "body harus berupa JSON yang valid")
	}

	if req.NIM == nil &&
		req.Name == nil &&
		req.Grade == nil &&
		req.IsActive == nil {
		return fail(c, fiber.StatusBadRequest, "tidak ada field yang diubah")
	}

	errs := map[string]string{}

	if req.NIM != nil && *req.NIM <= 0 {
		errs["nim"] = "harus berupa angka positif"
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)

		if name == "" {
			errs["name"] = "tidak boleh kosong"
		} else {
			*req.Name = name
		}
	}

	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 100 {
			errs["grade"] = "harus berada antara 0 sampai 100"
		}
	}

	if len(errs) > 0 {
		return failValidation(c, errs)
	}

	if req.NIM != nil {
		student.NIM = *req.NIM
	}

	if req.Name != nil {
		student.Name = *req.Name
	}

	if req.Grade != nil {
		student.Grade = *req.Grade
	}

	if req.IsActive != nil {
		student.IsActive = *req.IsActive
	}

	student, err = h.repo.Update(ctx, student)
	if err != nil {
		return handleRepositoryError(c, err)
	}

	return ok(c, "data mahasiswa berhasil diperbarui", student)
}

// Menghapus mahasiswa
func (h *StudentHandler) Delete(c *fiber.Ctx) error {
	id, valid := paramID(c)
	if !valid {
		return fail(c, fiber.StatusBadRequest, "id harus berupa angka positif")
	}

	ctx, cancel := requestContext(c)
	defer cancel()

	if err := h.repo.Delete(ctx, id); err != nil {
		return handleRepositoryError(c, err)
	}

	return noContent(c)
}
