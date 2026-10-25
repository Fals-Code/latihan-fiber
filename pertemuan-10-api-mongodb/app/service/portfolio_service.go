package service

import (
	"context"
	"encoding/hex"
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"tugas1-go/pertemuan-10-api-mongodb/app/model"
	"tugas1-go/pertemuan-10-api-mongodb/app/repository"
	"tugas1-go/pertemuan-10-api-mongodb/helper"
	p7model "tugas1-go/pertemuan-7-advanced-api-design/app/model"
	p7repository "tugas1-go/pertemuan-7-advanced-api-design/app/repository"
	p7helper "tugas1-go/pertemuan-7-advanced-api-design/helper"
)

type studentRepository interface {
	FindByID(context.Context, int) (p7model.Student, error)
}

// PortfolioService coordinates student ownership in PostgreSQL with portfolio operations in MongoDB.
type PortfolioService struct {
	portfolios  repository.PortfolioRepository
	students    studentRepository
	permissions *p7helper.PermissionSet
}

func NewPortfolioService(
	portfolios repository.PortfolioRepository,
	students studentRepository,
	permissions *p7helper.PermissionSet,
) *PortfolioService {
	return &PortfolioService{portfolios: portfolios, students: students, permissions: permissions}
}

func (s *PortfolioService) Get(c *fiber.Ctx) error {
	studentID, ok := portfolioStudentID(c)
	if !ok {
		return badStudentID()
	}
	ctx, cancel := p7helper.RequestContext(c)
	defer cancel()

	student, err := s.findStudent(ctx, studentID)
	if err != nil {
		return err
	}
	if !s.canAccess(c, student, "student:read:any") {
		return forbiddenStudent()
	}

	portfolio, err := s.portfolios.FindByStudentID(ctx, studentID)
	if err != nil {
		return translatePortfolioError(err)
	}
	return p7helper.Success(c, fiber.StatusOK, "portfolio berhasil diambil", portfolio)
}

func (s *PortfolioService) Replace(c *fiber.Ctx) error {
	request, err := decodePortfolioBody(c)
	if err != nil {
		return err
	}
	if fields := helper.ValidatePortfolioRequest(request); len(fields) > 0 {
		return invalidPortfolioRequest(fields)
	}
	studentID, ok := portfolioStudentID(c)
	if !ok {
		return badStudentID()
	}
	ctx, cancel := p7helper.RequestContext(c)
	defer cancel()

	student, err := s.findStudent(ctx, studentID)
	if err != nil {
		return err
	}
	if !s.canAccess(c, student, "student:update:any") {
		return forbiddenStudent()
	}

	portfolio, created, err := s.portfolios.Replace(ctx, studentID, request)
	if err != nil {
		return translatePortfolioError(err)
	}
	if created {
		return p7helper.Created(c, "portfolio berhasil dibuat", portfolio, "/api/v1/students/"+strconv.Itoa(studentID)+"/portfolio")
	}
	return p7helper.Success(c, fiber.StatusOK, "portfolio berhasil diganti", portfolio)
}

func (s *PortfolioService) AddProject(c *fiber.Ctx) error {
	request, err := decodeProjectBody(c)
	if err != nil {
		return err
	}
	if fields := helper.ValidateProjectRequest(request); len(fields) > 0 {
		return invalidPortfolioRequest(fields)
	}
	studentID, ok := portfolioStudentID(c)
	if !ok {
		return badStudentID()
	}
	ctx, cancel := p7helper.RequestContext(c)
	defer cancel()

	student, err := s.findStudent(ctx, studentID)
	if err != nil {
		return err
	}
	if !s.canAccess(c, student, "student:update:any") {
		return forbiddenStudent()
	}

	project, err := s.portfolios.AddProject(ctx, studentID, request)
	if err != nil {
		return translatePortfolioError(err)
	}
	return p7helper.Created(c, "project berhasil ditambahkan", project, "/api/v1/students/"+strconv.Itoa(studentID)+"/portfolio/projects/"+project.ID)
}

func (s *PortfolioService) DeleteProject(c *fiber.Ctx) error {
	studentID, ok := portfolioStudentID(c)
	if !ok {
		return badStudentID()
	}
	projectID := c.Params("projectId")
	if !validProjectID(projectID) {
		return p7helper.NewAppError(fiber.StatusBadRequest, "projectId tidak valid")
	}
	ctx, cancel := p7helper.RequestContext(c)
	defer cancel()

	student, err := s.findStudent(ctx, studentID)
	if err != nil {
		return err
	}
	if !s.canAccess(c, student, "student:update:any") {
		return forbiddenStudent()
	}
	if err := s.portfolios.DeleteProject(ctx, studentID, projectID); err != nil {
		return translatePortfolioError(err)
	}
	return p7helper.NoContent(c)
}

func validProjectID(projectID string) bool {
	decoded, err := hex.DecodeString(projectID)
	return err == nil && len(decoded) == 12
}

func (s *PortfolioService) TechnologyStats(c *fiber.Ctx) error {
	ctx, cancel := p7helper.RequestContext(c)
	defer cancel()
	statistics, err := s.portfolios.TopTechnologies(ctx)
	if err != nil {
		return translatePortfolioError(err)
	}
	return p7helper.Success(c, fiber.StatusOK, "statistik teknologi berhasil diambil", statistics)
}

func (s *PortfolioService) findStudent(ctx context.Context, studentID int) (p7model.Student, error) {
	student, err := s.students.FindByID(ctx, studentID)
	if errors.Is(err, p7repository.ErrNotFound) {
		return p7model.Student{}, p7helper.NewAppError(fiber.StatusNotFound, "mahasiswa tidak ditemukan")
	}
	if err != nil {
		return p7model.Student{}, p7helper.Internal(err)
	}
	return student, nil
}

func (s *PortfolioService) canAccess(c *fiber.Ctx, student p7model.Student, permission string) bool {
	current, ok := p7helper.CurrentUser(c)
	if !ok {
		return false
	}
	if student.OwnerID != nil {
		if current.UserID == *student.OwnerID {
			return true
		}
	}
	return s.permissions != nil && s.permissions.Can(current.Role, permission)
}

func portfolioStudentID(c *fiber.Ctx) (int, bool) {
	id, err := strconv.Atoi(c.Params("id"))
	return id, err == nil && id > 0
}

func badStudentID() error {
	return p7helper.NewAppError(fiber.StatusBadRequest, "id harus berupa angka positif")
}

func forbiddenStudent() error {
	return p7helper.NewAppError(fiber.StatusForbidden, "tidak memiliki akses ke mahasiswa ini")
}

func invalidPortfolioRequest(fields map[string]string) error {
	return p7helper.NewAppError(fiber.StatusBadRequest, "validasi request gagal", fields)
}

func translatePortfolioError(err error) error {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return p7helper.NewAppError(fiber.StatusNotFound, "portfolio atau project tidak ditemukan")
	case errors.Is(err, repository.ErrProjectLimit):
		return p7helper.NewAppError(fiber.StatusConflict, "jumlah project maksimal 20")
	case errors.Is(err, repository.ErrInvalidID):
		return p7helper.NewAppError(fiber.StatusBadRequest, "projectId tidak valid")
	case errors.Is(err, repository.ErrDuplicate):
		return p7helper.NewAppError(fiber.StatusConflict, "portfolio mahasiswa sudah tersedia")
	case errors.Is(err, repository.ErrUnavailable):
		return &p7helper.AppError{Status: fiber.StatusServiceUnavailable, Code: "SERVICE_UNAVAILABLE", Message: "MongoDB tidak tersedia"}
	default:
		return p7helper.Internal(err)
	}
}

func decodePortfolioBody(c *fiber.Ctx) (model.PortfolioRequest, error) {
	request, err := helper.DecodePortfolioRequest(c.Body())
	if err != nil {
		return model.PortfolioRequest{}, p7helper.NewAppError(fiber.StatusBadRequest, "body harus berupa JSON yang valid dan hanya memuat field portfolio")
	}
	return request, nil
}

func decodeProjectBody(c *fiber.Ctx) (model.ProjectRequest, error) {
	request, err := helper.DecodeProjectRequest(c.Body())
	if err != nil {
		return model.ProjectRequest{}, p7helper.NewAppError(fiber.StatusBadRequest, "body harus berupa JSON yang valid dan hanya memuat field project")
	}
	return request, nil
}
