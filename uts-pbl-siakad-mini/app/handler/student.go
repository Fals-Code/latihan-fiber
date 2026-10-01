package handler

import (
	"errors"
	"log"
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"tugas1-go/uts-pbl-siakad-mini/app/dto"
	"tugas1-go/uts-pbl-siakad-mini/app/model"
	"tugas1-go/uts-pbl-siakad-mini/app/service"
	"tugas1-go/uts-pbl-siakad-mini/middleware"
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

func (h *StudentHandler) Detail(c *fiber.Ctx) error {
	id, err := parseStudentID(c.Params("id"))
	if err != nil {
		return validationFailure(c, map[string][]string{"id": {"ID mahasiswa harus berupa bilangan bulat positif"}})
	}
	actor, ok := c.Locals(middleware.AuthUserLocal).(model.CurrentUser)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(dto.ErrorResponse{Success: false, Message: "Token tidak valid atau akun tidak aktif"})
	}
	detail, err := h.service.Detail(c.UserContext(), id, actor)
	if errors.Is(err, service.ErrStudentNotFound) {
		return studentNotFound(c)
	}
	if errors.Is(err, service.ErrStudentForbidden) {
		return c.Status(fiber.StatusForbidden).JSON(dto.ErrorResponse{Success: false, Message: "Akses ditolak"})
	}
	if err != nil {
		log.Printf("student detail failed: %v", err)
		return internalFailure(c)
	}
	return c.Status(fiber.StatusOK).JSON(dto.Envelope{Success: true, Message: "Data mahasiswa berhasil diambil", Data: studentDetailResponse(detail)})
}

func (h *StudentHandler) Update(c *fiber.Ctx) error {
	id, err := parseStudentID(c.Params("id"))
	if err != nil {
		return validationFailure(c, map[string][]string{"id": {"ID mahasiswa harus berupa bilangan bulat positif"}})
	}
	var request map[string]any
	if err := c.BodyParser(&request); err != nil {
		return validationFailure(c, map[string][]string{"body": {"Body harus berupa JSON yang valid"}})
	}
	student, validation, err := h.service.Update(c.UserContext(), id, request)
	if validation != nil {
		return validationFailure(c, validation)
	}
	if errors.Is(err, service.ErrStudentNotFound) {
		return studentNotFound(c)
	}
	if err != nil {
		log.Printf("student update failed: %v", err)
		return internalFailure(c)
	}
	return c.Status(fiber.StatusOK).JSON(dto.Envelope{Success: true, Message: "Data mahasiswa berhasil diperbarui", Data: dto.StudentResponse{ID: student.ID, NIM: student.NIM, Nama: student.Nama, Prodi: student.Prodi, Angkatan: student.Angkatan, IPKTerakhir: student.IPKTerakhir}})
}

func (h *StudentHandler) Delete(c *fiber.Ctx) error {
	id, err := parseStudentID(c.Params("id"))
	if err != nil {
		return validationFailure(c, map[string][]string{"id": {"ID mahasiswa harus berupa bilangan bulat positif"}})
	}
	if err := h.service.Delete(c.UserContext(), id); errors.Is(err, service.ErrStudentNotFound) {
		return studentNotFound(c)
	} else if err != nil {
		log.Printf("student delete failed: %v", err)
		return internalFailure(c)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func parseStudentID(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid student ID")
	}
	return id, nil
}

func studentDetailResponse(detail model.StudentDetail) dto.StudentDetailResponse {
	courses := make([]dto.StudentCourseResponse, 0, len(detail.Courses))
	for _, course := range detail.Courses {
		courses = append(courses, dto.StudentCourseResponse{EnrollmentID: course.EnrollmentID, CourseID: course.CourseID, KodeMK: course.KodeMK, NamaMK: course.NamaMK, SKS: course.SKS, Semester: course.Semester, TahunAkademik: course.TahunAkademik})
	}
	return dto.StudentDetailResponse{ID: detail.ID, NIM: detail.NIM, Nama: detail.Nama, Prodi: detail.Prodi, Angkatan: detail.Angkatan, IPKTerakhir: detail.IPKTerakhir, Courses: courses, TotalSKS: detail.TotalSKS, BatasSKS: service.SKSLimit(detail.IPKTerakhir)}
}

func studentNotFound(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{Success: false, Message: "Mahasiswa tidak ditemukan"})
}

func validationFailure(c *fiber.Ctx, validation map[string][]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(dto.ErrorResponse{Success: false, Message: "Validasi gagal", Errors: validation})
}

func internalFailure(c *fiber.Ctx) error {
	return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{Success: false, Message: "Terjadi kesalahan pada server"})
}
