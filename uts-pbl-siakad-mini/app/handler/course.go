package handler

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"tugas1-go/uts-pbl-siakad-mini/app/dto"
	"tugas1-go/uts-pbl-siakad-mini/app/service"
)

type CourseHandler struct{ service *service.CourseService }

func NewCourseHandler(courseService *service.CourseService) *CourseHandler {
	return &CourseHandler{service: courseService}
}

func (h *CourseHandler) List(c *fiber.Ctx) error {
	courses, validation, err := h.service.List(c.UserContext(), map[string]string{
		"semester":  c.Query("semester"),
		"search":    c.Query("search"),
		"available": c.Query("available"),
	})
	if validation != nil {
		return validationFailure(c, validation)
	}
	if err != nil {
		log.Printf("course list failed: %v", err)
		return internalFailure(c)
	}
	data := make([]dto.CourseResponse, 0, len(courses))
	for _, course := range courses {
		data = append(data, dto.CourseResponse{ID: course.ID, KodeMK: course.KodeMK, NamaMK: course.NamaMK, SKS: course.SKS, Semester: course.Semester, Kuota: course.Kuota, Terisi: course.Terisi, SisaKuota: course.SisaKuota})
	}
	return c.Status(fiber.StatusOK).JSON(dto.Envelope{Success: true, Message: "Data mata kuliah berhasil diambil", Data: data})
}
