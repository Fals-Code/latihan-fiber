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
func (routeStudentStore) GetDetail(context.Context, int64) (model.StudentDetail, error) {
	return model.StudentDetail{}, nil
}
func (routeStudentStore) StudentIDByUserID(context.Context, int64) (int64, error) { return 0, nil }
func (routeStudentStore) Update(context.Context, int64, model.StudentUpdate) (model.Student, error) {
	return model.Student{}, nil
}
func (routeStudentStore) SoftDelete(context.Context, int64) error { return nil }

type routeEnrollmentStore struct{}

func (routeEnrollmentStore) Create(context.Context, int64, model.NewEnrollment) (model.Enrollment, error) {
	return model.Enrollment{}, nil
}
func (routeEnrollmentStore) Delete(context.Context, int64, int64) error { return nil }

type routeCourseStore struct{}

func (routeCourseStore) List(context.Context, model.CourseFilters) ([]model.Course, error) {
	return []model.Course{}, nil
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
	courseHandler := handler.NewCourseHandler(service.NewCourseService(routeCourseStore{}))
	enrollmentHandler := handler.NewEnrollmentHandler(service.NewEnrollmentService(routeEnrollmentStore{}))
	app := fiber.New()
	Register(app, authHandler, studentHandler, courseHandler, enrollmentHandler, []byte("route-test-secret-long-enough"), store)
	for _, endpoint := range []struct{ method, path string }{
		{"POST", "/api/v1/auth/login"}, {"GET", "/api/v1/auth/me"}, {"GET", "/api/v1/students"}, {"POST", "/api/v1/students"}, {"GET", "/api/v1/students/5"}, {"PUT", "/api/v1/students/5"}, {"DELETE", "/api/v1/students/5"},
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
	for _, path := range []string{"/api/v1/courses", "/api/v1/enrollments"} {
		request := httptest.NewRequest("GET", path, nil)
		request.Header.Set("Authorization", "Bearer "+loginBody.Data.AccessToken)
		response, err := app.Test(request, -1)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if path == "/api/v1/enrollments" && response.StatusCode != fiber.StatusForbidden {
			t.Fatalf("unexpected route %s status: %d", path, response.StatusCode)
		}
		if path == "/api/v1/courses" && response.StatusCode != fiber.StatusOK {
			t.Fatalf("unexpected route %s status: %d", path, response.StatusCode)
		}
	}
}

func TestEnrollmentRoutesRequireStudentRoleAndAuthentication(t *testing.T) {
	secret := []byte("route-test-secret-long-enough")
	mkApp := func(role string) (*fiber.App, routeAuthStore) {
		hash, _ := bcrypt.GenerateFromPassword([]byte("password-123"), bcrypt.MinCost)
		store := routeAuthStore{user: model.User{ID: 9, Email: "person@example.com", Password: string(hash), Role: role}}
		authHandler := handler.NewAuthHandler(service.NewAuthService(store, string(secret), time.Hour), store)
		students := handler.NewStudentHandler(service.NewStudentService(routeStudentStore{}))
		courses := handler.NewCourseHandler(service.NewCourseService(routeCourseStore{}))
		enrollments := handler.NewEnrollmentHandler(service.NewEnrollmentService(routeEnrollmentStore{}))
		app := fiber.New()
		Register(app, authHandler, students, courses, enrollments, secret, store)
		return app, store
	}
	for _, endpoint := range []struct{ method, path string }{{"POST", "/api/v1/enrollments"}, {"DELETE", "/api/v1/enrollments/1"}} {
		adminApp, adminStore := mkApp("admin")
		adminAuth := service.NewAuthService(adminStore, string(secret), time.Hour)
		adminToken, _ := adminAuth.Login(context.Background(), "person@example.com", "password-123")
		request := httptest.NewRequest(endpoint.method, endpoint.path, strings.NewReader(`{"course_id":1,"tahun_akademik":"2026/2027-Ganjil"}`))
		request.Header.Set("Authorization", "Bearer "+adminToken.Token)
		response, err := adminApp.Test(request, -1)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != fiber.StatusForbidden {
			t.Errorf("admin %s %s status=%d", endpoint.method, endpoint.path, response.StatusCode)
		}

		studentApp, studentStore := mkApp("mahasiswa")
		studentAuth := service.NewAuthService(studentStore, string(secret), time.Hour)
		studentToken, _ := studentAuth.Login(context.Background(), "person@example.com", "password-123")
		studentRequest := httptest.NewRequest(endpoint.method, endpoint.path, strings.NewReader(`{"course_id":1,"tahun_akademik":"2026/2027-Ganjil"}`))
		studentRequest.Header.Set("Authorization", "Bearer "+studentToken.Token)
		studentResponse, err := studentApp.Test(studentRequest, -1)
		if err != nil {
			t.Fatal(err)
		}
		studentResponse.Body.Close()
		if studentResponse.StatusCode == fiber.StatusForbidden || studentResponse.StatusCode == fiber.StatusUnauthorized {
			t.Errorf("student %s %s status=%d", endpoint.method, endpoint.path, studentResponse.StatusCode)
		}

		unauthenticated, err := adminApp.Test(httptest.NewRequest(endpoint.method, endpoint.path, strings.NewReader(`{}`)), -1)
		if err != nil {
			t.Fatal(err)
		}
		unauthenticated.Body.Close()
		if unauthenticated.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("unauthenticated %s %s status=%d", endpoint.method, endpoint.path, unauthenticated.StatusCode)
		}
	}
}
