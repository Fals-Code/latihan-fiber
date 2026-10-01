package service

import (
	"context"
	"errors"
	"testing"

	"tugas1-go/uts-pbl-siakad-mini/app/model"
)

type courseTestStore struct {
	filters model.CourseFilters
	courses []model.Course
	err     error
}

func (s *courseTestStore) List(_ context.Context, filters model.CourseFilters) ([]model.Course, error) {
	s.filters = filters
	return s.courses, s.err
}

func TestCourseListValidatesAndPassesCombinedFilters(t *testing.T) {
	store := &courseTestStore{courses: []model.Course{}}
	service := NewCourseService(store)
	courses, validation, err := service.List(context.Background(), map[string]string{"semester": "3", "search": "Web", "available": "true"})
	if err != nil || validation != nil || len(courses) != 0 {
		t.Fatalf("List() = (%v, %v, %v)", courses, validation, err)
	}
	if store.filters.Semester == nil || *store.filters.Semester != 3 || store.filters.Search != "Web" || !store.filters.Available {
		t.Fatalf("filters not propagated: %+v", store.filters)
	}
}

func TestCourseListRejectsSemesterOutsideSchemaRange(t *testing.T) {
	store := &courseTestStore{}
	_, validation, err := NewCourseService(store).List(context.Background(), map[string]string{"semester": "15"})
	if err != nil || validation["semester"] == nil {
		t.Fatalf("expected semester validation, got %v, %v", validation, err)
	}
}

func TestCourseListRejectsInvalidAvailability(t *testing.T) {
	_, validation, err := NewCourseService(&courseTestStore{}).List(context.Background(), map[string]string{"available": "yes"})
	if err != nil || validation["available"] == nil {
		t.Fatalf("expected available validation, got %v, %v", validation, err)
	}
}

func TestCourseListReturnsRepositoryError(t *testing.T) {
	want := errors.New("database failed")
	_, validation, err := NewCourseService(&courseTestStore{err: want}).List(context.Background(), nil)
	if validation != nil || !errors.Is(err, want) {
		t.Fatalf("List() error = %v, validation = %v", err, validation)
	}
}
