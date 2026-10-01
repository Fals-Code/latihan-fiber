package route

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"

	"tugas1-go/uts-pbl-siakad-mini/app/handler"
	"tugas1-go/uts-pbl-siakad-mini/app/model"
	"tugas1-go/uts-pbl-siakad-mini/app/service"
)

type routeAuthStore struct{ user model.User }

func (s routeAuthStore) FindUserByEmail(context.Context, string) (model.User, error) {
	return s.user, nil
}
func (s routeAuthStore) FindActiveAccount(_ context.Context, id int64) (model.CurrentUser, error) {
	return model.CurrentUser{ID: id, Email: s.user.Email, Role: s.user.Role}, nil
}

type routeStudentStore struct{}

func (routeStudentStore) List(context.Context, model.StudentFilters) ([]model.Student, int64, error) {
	return []model.Student{}, 0, nil
}
func (routeStudentStore) Create(context.Context, model.NewStudent, func() (string, error)) (model.Student, error) {
	return model.Student{}, nil
}

func TestRegisterOnlyCurrentEndpoints(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password-123"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	store := routeAuthStore{user: model.User{ID: 1, Email: "admin@example.com", Password: string(hash), Role: "admin"}}
	authService := service.NewAuthService(store, "route-test-secret-long-enough", time.Hour)
	authHandler := handler.NewAuthHandler(authService, store)
	studentService := service.NewStudentService(routeStudentStore{})
	studentHandler := handler.NewStudentHandler(studentService)
	app := fiber.New()
	Register(app, authHandler, studentHandler, []byte("route-test-secret-long-enough"), store)
	for _, endpoint := range []struct{ method, path string }{
		{"POST", "/api/v1/auth/login"}, {"GET", "/api/v1/auth/me"}, {"GET", "/api/v1/students"}, {"POST", "/api/v1/students"},
	} {
		request := httptest.NewRequest(endpoint.method, endpoint.path, nil)
		response, err := app.Test(request, -1)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode == 404 {
			t.Fatalf("route %s %s is missing", endpoint.method, endpoint.path)
		}
	}
	loginRequest := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{"email":"admin@example.com","password":"password-123"}`))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse, err := app.Test(loginRequest, -1)
	if err != nil {
		t.Fatal(err)
	}
	var loginBody struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(loginResponse.Body).Decode(&loginBody); err != nil {
		t.Fatal(err)
	}
	loginResponse.Body.Close()
	for _, path := range []string{"/api/v1/students/1", "/api/v1/courses", "/api/v1/enrollments"} {
		request := httptest.NewRequest("GET", path, nil)
		request.Header.Set("Authorization", "Bearer "+loginBody.Data.AccessToken)
		response, err := app.Test(request, -1)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != 404 {
			t.Fatalf("unexpected route %s status: %d", path, response.StatusCode)
		}
	}
}
