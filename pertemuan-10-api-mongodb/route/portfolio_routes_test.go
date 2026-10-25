package route_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"tugas1-go/pertemuan-10-api-mongodb/app/model"
	"tugas1-go/pertemuan-10-api-mongodb/app/repository"
	"tugas1-go/pertemuan-10-api-mongodb/app/service"
	"tugas1-go/pertemuan-10-api-mongodb/route"
	p7model "tugas1-go/pertemuan-7-advanced-api-design/app/model"
	p7helper "tugas1-go/pertemuan-7-advanced-api-design/helper"
)

type studentsFake struct{ student p7model.Student }

func (f studentsFake) FindByID(context.Context, int) (p7model.Student, error) {
	return f.student, nil
}

type portfoliosFake struct{}

func (portfoliosFake) EnsureIndexes(context.Context) error { return nil }
func (portfoliosFake) FindByStudentID(context.Context, int) (model.Portfolio, error) {
	return model.Portfolio{StudentID: 1, Headline: "Test portfolio", Projects: []model.Project{}, Certificates: []model.Certificate{}}, nil
}
func (portfoliosFake) Replace(context.Context, int, model.PortfolioRequest) (model.Portfolio, bool, error) {
	return model.Portfolio{}, false, nil
}
func (portfoliosFake) AddProject(context.Context, int, model.ProjectRequest) (model.Project, error) {
	return model.Project{}, nil
}
func (portfoliosFake) DeleteProject(context.Context, int, string) error { return nil }
func (portfoliosFake) TopTechnologies(context.Context) ([]model.TechnologyCount, error) {
	return []model.TechnologyCount{}, nil
}

func TestRegisterPortfolioRoutesUsesAPIV1AndJWT(t *testing.T) {
	ownerID := 3
	permissions := p7helper.NewPermissionSet(nil)
	portfolioService := service.NewPortfolioService(
		portfoliosFake{},
		studentsFake{student: p7model.Student{ID: 1, OwnerID: &ownerID}},
		permissions,
	)
	jwt := p7helper.NewJWTManager("test-secret-long-enough-for-phase-three", "test-issuer", time.Minute)
	app := fiber.New(fiber.Config{ErrorHandler: func(c *fiber.Ctx, err error) error {
		var appErr *p7helper.AppError
		if errors.As(err, &appErr) {
			return c.Status(appErr.Status).JSON(map[string]any{"success": false, "code": appErr.Code})
		}
		return c.SendStatus(http.StatusInternalServerError)
	}})
	route.RegisterPortfolioRoutes(app.Group("/api/v1"), portfolioService, jwt)

	unauthorized := httptest.NewRequest(http.MethodGet, "/api/v1/students/1/portfolio", nil)
	unauthorizedResponse, err := app.Test(unauthorized)
	if err != nil {
		t.Fatal(err)
	}
	if unauthorizedResponse.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthorized status=%d want=%d", unauthorizedResponse.StatusCode, http.StatusUnauthorized)
	}

	token, err := jwt.GenerateAccessToken(p7model.User{ID: ownerID, Username: "owner", Role: "user"})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/students/1/portfolio", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("authorized status=%d want=%d", response.StatusCode, http.StatusOK)
	}
}

var _ repository.PortfolioRepository = portfoliosFake{}
