package handler

import (
	"context"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	"tugas1-go/uts-pbl-siakad-mini/app/model"
	"tugas1-go/uts-pbl-siakad-mini/app/repository"
	"tugas1-go/uts-pbl-siakad-mini/app/service"
	"tugas1-go/uts-pbl-siakad-mini/middleware"
)

type enrollmentHandlerStore struct{ createErr, deleteErr error }

func (s enrollmentHandlerStore) Create(_ context.Context, _ int64, input model.NewEnrollment) (model.Enrollment, error) {
	return model.Enrollment{ID: 1, CourseID: input.CourseID, TahunAkademik: input.TahunAkademik}, s.createErr
}
func (s enrollmentHandlerStore) Delete(context.Context, int64, int64) error { return s.deleteErr }

func enrollmentTestApp(store enrollmentHandlerStore) *fiber.App {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(middleware.AuthUserLocal, model.CurrentUser{ID: 9, Role: "mahasiswa"})
		return c.Next()
	})
	handler := NewEnrollmentHandler(service.NewEnrollmentService(store))
	app.Post("/enrollments", handler.Create)
	app.Delete("/enrollments/:id", handler.Delete)
	return app
}

func TestEnrollmentCreateRejectsUnknownFields(t *testing.T) {
	response, err := enrollmentTestApp(enrollmentHandlerStore{}).Test(httptest.NewRequest("POST", "/enrollments", strings.NewReader(`{"course_id":1,"tahun_akademik":"2025/2026-Ganjil","student_id":99}`)))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != fiber.StatusUnprocessableEntity {
		t.Fatalf("status=%d", response.StatusCode)
	}
}

func TestEnrollmentCreateMapsBusinessErrors(t *testing.T) {
	for _, test := range []struct {
		name   string
		cause  error
		status int
	}{
		{"duplicate", repository.ErrDuplicateEnrollment, fiber.StatusConflict},
		{"quota", repository.ErrEnrollmentQuotaFull, fiber.StatusUnprocessableEntity},
		{"sks", fmt.Errorf("%w: sisa SKS yang dapat diambil 3", repository.ErrEnrollmentSKSLimit), fiber.StatusUnprocessableEntity},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/enrollments", strings.NewReader(`{"course_id":1,"tahun_akademik":"2025/2026-Ganjil"}`))
			request.Header.Set("Content-Type", "application/json")
			response, err := enrollmentTestApp(enrollmentHandlerStore{createErr: test.cause}).Test(request)
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != test.status {
				t.Fatalf("status=%d want=%d", response.StatusCode, test.status)
			}
		})
	}
}

func TestEnrollmentDeleteMapsOwnership(t *testing.T) {
	for _, test := range []struct {
		name   string
		cause  error
		status int
	}{
		{"missing", repository.ErrNotFound, fiber.StatusNotFound},
		{"other student", repository.ErrForbidden, fiber.StatusForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			response, err := enrollmentTestApp(enrollmentHandlerStore{deleteErr: test.cause}).Test(httptest.NewRequest("DELETE", "/enrollments/3", nil))
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			if response.StatusCode != test.status {
				t.Fatalf("status=%d want=%d", response.StatusCode, test.status)
			}
		})
	}
}
