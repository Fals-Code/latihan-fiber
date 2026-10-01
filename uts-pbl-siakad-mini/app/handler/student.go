package handler

import (
	"log"
	"math"

	"github.com/gofiber/fiber/v2"

	"tugas1-go/uts-pbl-siakad-mini/app/dto"
	"tugas1-go/uts-pbl-siakad-mini/app/model"
	"tugas1-go/uts-pbl-siakad-mini/app/service"
)

type StudentHandler struct{ service *service.StudentService }

func NewStudentHandler(studentService *service.StudentService) *StudentHandler {
	return &StudentHandler{service: studentService}
}

func (h *StudentHandler) List(c *fiber.Ctx) error {
	query := map[string]string{
		"page": c.Query("page"), "per_page": c.Query("per_page"), "prodi": c.Query("prodi"),
		"angkatan": c.Query("angkatan"), "search": c.Query("search"), "sort": c.Query("sort"),
	}
	students, filters, total, validation, err := h.service.List(c.UserContext(), query)
	if validation != nil {
		return validationFailure(c, validation)
	}
	if err != nil {
		log.Printf("student list failed: %v", err)
		return internalFailure(c)
	}
	data := make([]dto.StudentResponse, 0, len(students))
	for _, student := range students {
		data = append(data, dto.StudentResponse{ID: student.ID, NIM: student.NIM, Nama: student.Nama, Prodi: student.Prodi, Angkatan: student.Angkatan, IPKTerakhir: student.IPKTerakhir})
	}
	lastPage := 0
	if total > 0 {
		lastPage = int(math.Ceil(float64(total) / float64(filters.PerPage)))
	}
	return c.Status(fiber.StatusOK).JSON(dto.StudentListEnvelope{Success: true, Message: "Data mahasiswa berhasil diambil", Data: data, Meta: dto.PaginationMeta{CurrentPage: filters.Page, PerPage: filters.PerPage, Total: total, LastPage: lastPage}})
}

func (h *StudentHandler) Create(c *fiber.Ctx) error {
	var request dto.CreateStudentRequest
	if err := c.BodyParser(&request); err != nil {
		return validationFailure(c, map[string][]string{"body": {"Body harus berupa JSON yang valid"}})
	}
	if request.Angkatan == nil {
		return validationFailure(c, map[string][]string{"angkatan": {"Angkatan wajib diisi"}})
	}
	ipk := 0.0
	if request.IPKTerkahir != nil {
		ipk = *request.IPKTerkahir
	}
	student, validation, err := h.service.Create(c.UserContext(), model.NewStudent{NIM: request.NIM, Nama: request.Nama, Email: request.Email, Prodi: request.Prodi, Angkatan: *request.Angkatan, IPKTerakhir: ipk})
	if validation != nil {
		return validationFailure(c, validation)
	}
	if err != nil {
		log.Printf("student create failed: %v", err)
		return internalFailure(c)
	}
	return c.Status(fiber.StatusCreated).JSON(dto.Envelope{Success: true, Message: "Mahasiswa berhasil dibuat", Data: dto.StudentResponse{ID: student.ID, NIM: student.NIM, Nama: student.Nama, Prodi: student.Prodi, Angkatan: student.Angkatan, IPKTerakhir: student.IPKTerakhir}})
}

func validationFailure(c *fiber.Ctx, validation map[string][]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(dto.ErrorResponse{Success: false, Message: "Validasi gagal", Errors: validation})
}

func internalFailure(c *fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Success: false, Message: "Terjadi kesalahan pada server"})
}
