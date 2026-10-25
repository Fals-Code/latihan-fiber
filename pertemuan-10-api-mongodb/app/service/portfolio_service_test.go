package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"tugas1-go/pertemuan-10-api-mongodb/app/model"
	"tugas1-go/pertemuan-10-api-mongodb/app/repository"
	p7model "tugas1-go/pertemuan-7-advanced-api-design/app/model"
	p7repository "tugas1-go/pertemuan-7-advanced-api-design/app/repository"
	p7helper "tugas1-go/pertemuan-7-advanced-api-design/helper"
	p7middleware "tugas1-go/pertemuan-7-advanced-api-design/middleware"
)

type portfolioRepoFake struct {
	portfolio model.Portfolio
	project   model.Project
	stats     []model.TechnologyCount
	created   bool
	findErr   error
	writeErr  error
	deleteErr error
	statsErr  error
	calls     int
}

func (f *portfolioRepoFake) EnsureIndexes(context.Context) error { return nil }
func (f *portfolioRepoFake) FindByStudentID(context.Context, int) (model.Portfolio, error) {
	f.calls++
	return f.portfolio, f.findErr
}
func (f *portfolioRepoFake) Replace(_ context.Context, studentID int, request model.PortfolioRequest) (model.Portfolio, bool, error) {
	f.calls++
	return model.Portfolio{StudentID: studentID, Headline: request.Headline, Bio: request.Bio, Projects: []model.Project{}, Certificates: []model.Certificate{}}, f.created, f.writeErr
}
func (f *portfolioRepoFake) AddProject(context.Context, int, model.ProjectRequest) (model.Project, error) {
	f.calls++
	return f.project, f.writeErr
}
func (f *portfolioRepoFake) DeleteProject(context.Context, int, string) error {
	f.calls++
	return f.deleteErr
}
func (f *portfolioRepoFake) TopTechnologies(context.Context) ([]model.TechnologyCount, error) {
	f.calls++
	return f.stats, f.statsErr
}

type studentRepoFake struct {
	student p7model.Student
	err     error
	calls   int
}

func (f *studentRepoFake) FindByID(context.Context, int) (p7model.Student, error) {
	f.calls++
	return f.student, f.err
}
func (f *studentRepoFake) FindOwnerIDByID(context.Context, int) (*int, error) {
	return f.student.OwnerID, f.err
}

func TestPortfolioRoutesRequireJWTAndOwnership(t *testing.T) {
	ownerID := 7
	repo := &portfolioRepoFake{}
	students := &studentRepoFake{student: p7model.Student{ID: 1, OwnerID: &ownerID}}
	app, jwt := newPortfolioTestApp(repo, students)

	request := httptest.NewRequest(http.MethodGet, "/api/v1/students/1/portfolio", nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusUnauthorized || repo.calls != 0 {
		t.Fatalf("without token: status=%d repository calls=%d", response.StatusCode, repo.calls)
	}
	response = performRequest(t, app, "invalid-token", http.MethodGet, "/api/v1/students/1/portfolio", "")
	if response.StatusCode != http.StatusUnauthorized || repo.calls != 0 {
		t.Fatalf("invalid token: status=%d repository calls=%d", response.StatusCode, repo.calls)
	}

	response = performRequest(t, app, jwtFor(t, jwt, 8), http.MethodGet, "/api/v1/students/1/portfolio", "")
	if response.StatusCode != http.StatusForbidden || repo.calls != 0 {
		t.Fatalf("non-owner: status=%d repository calls=%d", response.StatusCode, repo.calls)
	}

	response = performRequest(t, app, jwtFor(t, jwt, ownerID), http.MethodGet, "/api/v1/students/1/portfolio", "")
	if response.StatusCode != http.StatusOK || repo.calls != 1 {
		t.Fatalf("owner: status=%d repository calls=%d", response.StatusCode, repo.calls)
	}
}

func TestPortfolioPutCreatesAndValidatesBeforeRepository(t *testing.T) {
	ownerID := 7
	repo := &portfolioRepoFake{created: true}
	students := &studentRepoFake{student: p7model.Student{ID: 1, OwnerID: &ownerID}}
	app, jwt := newPortfolioTestApp(repo, students)
	token := jwtFor(t, jwt, ownerID)

	response := performRequest(t, app, token, http.MethodPut, "/api/v1/students/1/portfolio", `{"headline":"Engineer","bio":"Bio","projects":[],"certificates":[]}`)
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("valid PUT status=%d", response.StatusCode)
	}

	for _, body := range []string{
		`{"headline":"Engineer","bio":"Bio","projects":[],"certificates":[],"$set":{"bio":"injected"}}`,
		`{"headline":"Engineer","bio":"Bio","projects":[{"title":"x","technologies":["Go"],"link":"https://example.test","year":2025,"_id":{"$ne":null}}],"certificates":[]}`,
	} {
		before := repo.calls
		response = performRequest(t, app, token, http.MethodPut, "/api/v1/students/1/portfolio", body)
		if response.StatusCode != http.StatusBadRequest {
			t.Fatalf("malicious PUT status=%d body=%s", response.StatusCode, responseBody(t, response))
		}
		if repo.calls != before || students.calls != 1 {
			t.Fatalf("invalid request reached a repository: Mongo calls=%d before=%d; student calls=%d", repo.calls, before, students.calls)
		}
	}

	tooMany := `{"headline":"Engineer","bio":"Bio","projects":[` + strings.TrimSuffix(strings.Repeat(`{"title":"x","technologies":["Go"],"link":"https://example.test","year":2025},`, 21), ",") + `],"certificates":[]}`
	before := repo.calls
	response = performRequest(t, app, token, http.MethodPut, "/api/v1/students/1/portfolio", tooMany)
	if response.StatusCode != http.StatusBadRequest || repo.calls != before || students.calls != 1 {
		t.Fatalf("over-limit PUT status=%d Mongo calls=%d before=%d; student calls=%d", response.StatusCode, repo.calls, before, students.calls)
	}

	repo.created = false
	response = performRequest(t, app, token, http.MethodPut, "/api/v1/students/1/portfolio", `{"headline":"Replacement","bio":"Bio","projects":[],"certificates":[]}`)
	if response.StatusCode != http.StatusOK {
		t.Fatalf("valid replacement status=%d", response.StatusCode)
	}
}

func TestPortfolioMongoUnavailableReturns503(t *testing.T) {
	ownerID := 7
	repo := &portfolioRepoFake{findErr: repository.ErrUnavailable}
	app, jwt := newPortfolioTestApp(repo, &studentRepoFake{student: p7model.Student{ID: 1, OwnerID: &ownerID}})
	response := performRequest(t, app, jwtFor(t, jwt, ownerID), http.MethodGet, "/api/v1/students/1/portfolio", "")
	if response.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("status=%d want=%d", response.StatusCode, http.StatusServiceUnavailable)
	}
}

func TestPortfolioEndpointResponses(t *testing.T) {
	ownerID := 7
	projectID := "507f1f77bcf86cd799439011"
	repo := &portfolioRepoFake{
		portfolio: model.Portfolio{ID: "507f1f77bcf86cd799439012", StudentID: 1, Headline: "Engineer", Projects: []model.Project{}},
		project:   model.Project{ID: projectID, Title: "API", Technologies: []string{"Go"}, Link: "https://example.test", Year: 2025},
		stats:     []model.TechnologyCount{{Technology: "Go", Total: 1}},
	}
	app, jwt := newPortfolioTestApp(repo, &studentRepoFake{student: p7model.Student{ID: 1, OwnerID: &ownerID}})
	token := jwtFor(t, jwt, ownerID)

	cases := []struct {
		method, path, body string
		want               int
	}{
		{http.MethodGet, "/api/v1/students/1/portfolio", "", http.StatusOK},
		{http.MethodPost, "/api/v1/students/1/portfolio/projects", `{"title":"API","technologies":["Go"],"link":"https://example.test","year":2025}`, http.StatusCreated},
		{http.MethodDelete, "/api/v1/students/1/portfolio/projects/" + projectID, "", http.StatusNoContent},
		{http.MethodGet, "/api/v1/portfolios/stats/technologies", "", http.StatusOK},
	}
	for _, test := range cases {
		response := performRequest(t, app, token, test.method, test.path, test.body)
		if response.StatusCode != test.want {
			t.Errorf("%s %s status=%d want=%d body=%s", test.method, test.path, response.StatusCode, test.want, responseBody(t, response))
		}
	}
}

func TestPortfolioMissingStudentDoesNotCallMongo(t *testing.T) {
	repo := &portfolioRepoFake{}
	students := &studentRepoFake{err: p7repository.ErrNotFound}
	app, jwt := newPortfolioTestApp(repo, students)
	response := performRequest(t, app, jwtFor(t, jwt, 7), http.MethodGet, "/api/v1/students/1/portfolio", "")
	if response.StatusCode != http.StatusNotFound || repo.calls != 0 || students.calls != 1 {
		t.Fatalf("status=%d Mongo calls=%d student calls=%d", response.StatusCode, repo.calls, students.calls)
	}
}

func TestPortfolioRepositoryFailuresMapToHTTPResponses(t *testing.T) {
	ownerID := 7
	cases := []struct {
		name       string
		path       string
		method     string
		body       string
		repo       *portfolioRepoFake
		wantStatus int
	}{
		{
			name: "portfolio missing",
			path: "/api/v1/students/1/portfolio", method: http.MethodGet,
			repo: &portfolioRepoFake{findErr: repository.ErrNotFound}, wantStatus: http.StatusNotFound,
		},
		{
			name: "project limit reached",
			path: "/api/v1/students/1/portfolio/projects", method: http.MethodPost,
			body: `{"title":"API","technologies":["Go"],"link":"https://example.test","year":2025}`,
			repo: &portfolioRepoFake{writeErr: repository.ErrProjectLimit}, wantStatus: http.StatusConflict,
		},
		{
			name: "duplicate portfolio",
			path: "/api/v1/students/1/portfolio", method: http.MethodPut,
			body: `{"headline":"Engineer","bio":"Bio","projects":[],"certificates":[]}`,
			repo: &portfolioRepoFake{writeErr: repository.ErrDuplicate}, wantStatus: http.StatusConflict,
		},
		{
			name: "project missing",
			path: "/api/v1/students/1/portfolio/projects/507f1f77bcf86cd799439011", method: http.MethodDelete,
			repo: &portfolioRepoFake{deleteErr: repository.ErrNotFound}, wantStatus: http.StatusNotFound,
		},
		{
			name: "mongo unavailable",
			path: "/api/v1/portfolios/stats/technologies", method: http.MethodGet,
			repo: &portfolioRepoFake{statsErr: repository.ErrUnavailable}, wantStatus: http.StatusServiceUnavailable,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			app, jwt := newPortfolioTestApp(test.repo, &studentRepoFake{student: p7model.Student{ID: 1, OwnerID: &ownerID}})
			response := performRequest(t, app, jwtFor(t, jwt, ownerID), test.method, test.path, test.body)
			if response.StatusCode != test.wantStatus {
				t.Fatalf("status=%d want=%d response=%s", response.StatusCode, test.wantStatus, responseBody(t, response))
			}
		})
	}
}

func TestDeleteProjectRejectsInvalidIDBeforeDatabaseAccess(t *testing.T) {
	ownerID := 7
	repo := &portfolioRepoFake{}
	students := &studentRepoFake{student: p7model.Student{ID: 1, OwnerID: &ownerID}}
	app, jwt := newPortfolioTestApp(repo, students)
	response := performRequest(t, app, jwtFor(t, jwt, ownerID), http.MethodDelete,
		"/api/v1/students/1/portfolio/projects/not-an-object-id", "")
	if response.StatusCode != http.StatusBadRequest || repo.calls != 0 || students.calls != 0 {
		t.Fatalf("status=%d Mongo calls=%d student calls=%d", response.StatusCode, repo.calls, students.calls)
	}
}

func newPortfolioTestApp(repo repository.PortfolioRepository, students studentRepository) (*fiber.App, *p7helper.JWTManager) {
	jwt := p7helper.NewJWTManager("test-secret-long-enough-for-phase-three", "test-issuer", time.Minute)
	portfolioService := NewPortfolioService(repo, students, p7helper.NewPermissionSet(map[string][]string{"user": {}}))
	app := fiber.New(fiber.Config{ErrorHandler: func(c *fiber.Ctx, err error) error {
		var appErr *p7helper.AppError
		if errors.As(err, &appErr) {
			return c.Status(appErr.Status).JSON(map[string]any{"success": false, "code": appErr.Code})
		}
		return c.SendStatus(http.StatusInternalServerError)
	}})
	api := app.Group("/api/v1")
	requireAuth := p7middleware.RequireAuth(jwt)
	studentRoutes := api.Group("/students", p7middleware.RequireJSON, requireAuth)
	studentRoutes.Get("/:id/portfolio", portfolioService.Get)
	studentRoutes.Put("/:id/portfolio", portfolioService.Replace)
	studentRoutes.Post("/:id/portfolio/projects", portfolioService.AddProject)
	studentRoutes.Delete("/:id/portfolio/projects/:projectId", portfolioService.DeleteProject)
	api.Get("/portfolios/stats/technologies", requireAuth, portfolioService.TechnologyStats)
	return app, jwt
}

func jwtFor(t *testing.T, jwt *p7helper.JWTManager, userID int) string {
	t.Helper()
	token, err := jwt.GenerateAccessToken(p7model.User{ID: userID, Username: "p10-test", Role: "user"})
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func performRequest(t *testing.T, app *fiber.App, token, method, path, body string) *http.Response {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", fiber.MIMEApplicationJSON)
	}
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func responseBody(t *testing.T, response *http.Response) string {
	t.Helper()
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	return string(body)
}
