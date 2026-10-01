package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"

	"tugas1-go/uts-pbl-siakad-mini/app/model"
	"tugas1-go/uts-pbl-siakad-mini/app/service"
)

type fakeCourseServiceStore struct {
	courses []model.Course
	filters model.CourseFilters
	err     error
}

func (s *fakeCourseServiceStore) List(_ context.Context, filters model.CourseFilters) ([]model.Course, error) {
	s.filters = filters
	return s.courses, s.err
}

func TestCourseHandlerListReturnsAvailability(t *testing.T) {
	store := &fakeCourseServiceStore{courses: []model.Course{{ID: 7, KodeMK: "IF101", NamaMK: "Algoritma", SKS: 3, Semester: 1, Kuota: 30, Terisi: 12, SisaKuota: 18}}}
	app := fiber.New()
	app.Get("/courses", NewCourseHandler(service.NewCourseService(store)).List)
	response, err := app.Test(httptest.NewRequest("GET", "/courses", nil), -1)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var body struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusOK || len(body.Data) != 1 || body.Data[0]["terisi"] != float64(12) || body.Data[0]["sisa_kuota"] != float64(18) {
		t.Fatalf("unexpected response status=%d data=%v", response.StatusCode, body.Data)
	}
}

func TestCourseHandlerListRejectsInvalidQuery(t *testing.T) {
	store := &fakeCourseServiceStore{}
	app := fiber.New()
	app.Get("/courses", NewCourseHandler(service.NewCourseService(store)).List)
	response, err := app.Test(httptest.NewRequest("GET", "/courses?semester=99", nil), -1)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != fiber.StatusUnprocessableEntity {
		t.Fatalf("status=%d", response.StatusCode)
	}
}

func TestCourseHandlerListReturnsEmptyAndHidesInternalError(t *testing.T) {
	store := &fakeCourseServiceStore{courses: []model.Course{}}
	app := fiber.New()
	app.Get("/courses", NewCourseHandler(service.NewCourseService(store)).List)
	response, err := app.Test(httptest.NewRequest("GET", "/courses", nil), -1)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var empty struct {
		Data []model.Course `json:"data"`
	}
	if err := json.NewDecoder(response.Body).Decode(&empty); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusOK || empty.Data == nil || len(empty.Data) != 0 {
		t.Fatalf("empty response status=%d data=%v", response.StatusCode, empty.Data)
	}

	store.err = context.DeadlineExceeded
	response, err = app.Test(httptest.NewRequest("GET", "/courses", nil), -1)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var failure map[string]any
	if err := json.NewDecoder(response.Body).Decode(&failure); err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(failure)
	if response.StatusCode != fiber.StatusInternalServerError || string(body) == "" || string(body) == "context deadline exceeded" {
		t.Fatalf("internal error leaked or wrong status: %d %s", response.StatusCode, body)
	}
	if strings.Contains(string(body), "deadline") {
		t.Fatalf("internal error leaked: %s", body)
	}
}

func TestCourseHandlerListQueryValidation(t *testing.T) {
	store := &fakeCourseServiceStore{courses: []model.Course{}}
	app := fiber.New()
	app.Get("/courses", NewCourseHandler(service.NewCourseService(store)).List)
	for _, query := range []string{"semester=x", "semester=0", "semester=15", "available=yes"} {
		response, err := app.Test(httptest.NewRequest("GET", "/courses?"+query, nil), -1)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != fiber.StatusUnprocessableEntity {
			t.Errorf("query %q status=%d", query, response.StatusCode)
		}
	}
	for _, query := range []string{"semester=14&available=true", "available=false"} {
		response, err := app.Test(httptest.NewRequest("GET", "/courses?"+query, nil), -1)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != fiber.StatusOK {
			t.Errorf("query %q status=%d", query, response.StatusCode)
		}
	}
}
