package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"tugas1-go/uts-pbl-siakad-mini/app/dto"
	"tugas1-go/uts-pbl-siakad-mini/app/model"
	"tugas1-go/uts-pbl-siakad-mini/app/service"
	"tugas1-go/uts-pbl-siakad-mini/middleware"
)

type EnrollmentHandler struct{ service *service.EnrollmentService }

func NewEnrollmentHandler(enrollmentService *service.EnrollmentService) *EnrollmentHandler {
	return &EnrollmentHandler{service: enrollmentService}
}

func (h *EnrollmentHandler) Create(c *fiber.Ctx) error {
	var request dto.CreateEnrollmentRequest
	decoder := json.NewDecoder(bytes.NewReader(c.Body()))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return validationFailure(c, map[string][]string{"body": {"Body harus berupa JSON yang valid dan hanya memuat course_id serta tahun_akademik"}})
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return validationFailure(c, map[string][]string{"body": {"Body harus berisi satu objek JSON"}})
	}
	user, ok := c.Locals(middleware.AuthUserLocal).(model.CurrentUser)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Success: false, Message: "Token tidak valid atau akun tidak aktif"})
	}
	enrollment, validation, err := h.service.Create(c.UserContext(), user.ID, model.NewEnrollment{CourseID: request.CourseID, TahunAkademik: request.TahunAkademik})
	if validation != nil {
		return validationFailure(c, validation)
	}
	if errors.Is(err, service.ErrEnrollmentForbidden) {
		return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Success: false, Message: "Akses ditolak"})
	}
	if status, fields, handled := serviceEnrollmentBusinessError(err); handled {
		if status == fiber.StatusConflict {
			return c.Status(status).JSON(dto.ErrorResponse{Success: false, Message: "Mata kuliah sudah diambil pada tahun akademik tersebut"})
		}
		return validationFailure(c, fields)
	}
	if err != nil {
		log.Printf("enrollment create failed: %v", err)
		return internalFailure(c)
	}
	return c.Status(fiber.StatusCreated).JSON(dto.Envelope{Success: true, Message: "Mata kuliah berhasil diambil", Data: dto.EnrollmentResponse{ID: enrollment.ID, CourseID: enrollment.CourseID, TahunAkademik: enrollment.TahunAkademik, CreatedAt: enrollment.CreatedAt}})
}

func (h *EnrollmentHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil || id < 1 {
		return validationFailure(c, map[string][]string{"id": {"ID enrollment harus berupa bilangan bulat positif"}})
	}
	user, ok := c.Locals(middleware.AuthUserLocal).(model.CurrentUser)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Success: false, Message: "Token tidak valid atau akun tidak aktif"})
	}
	if err := h.service.Delete(c.UserContext(), user.ID, id); errors.Is(err, service.ErrEnrollmentNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Success: false, Message: "Enrollment tidak ditemukan"})
	} else if errors.Is(err, service.ErrEnrollmentForbidden) {
		return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Success: false, Message: "Akses ditolak"})
	} else if err != nil {
		log.Printf("enrollment delete failed: %v", err)
		return internalFailure(c)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func serviceEnrollmentBusinessError(err error) (int, map[string][]string, bool) {
	return service.EnrollmentBusinessError(err)
}
