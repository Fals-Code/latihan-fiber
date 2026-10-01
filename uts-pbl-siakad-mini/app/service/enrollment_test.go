package service

import (
	"context"
	"errors"
	"testing"

	"tugas1-go/uts-pbl-siakad-mini/app/model"
	"tugas1-go/uts-pbl-siakad-mini/app/repository"
)

type enrollmentTestStore struct {
	created   model.NewEnrollment
	createErr error
	deletedID int64
	deleteErr error
}

func (s *enrollmentTestStore) Create(_ context.Context, _ int64, input model.NewEnrollment) (model.Enrollment, error) {
	s.created = input
	return model.Enrollment{ID: 4, CourseID: input.CourseID, TahunAkademik: input.TahunAkademik}, s.createErr
}
func (s *enrollmentTestStore) Delete(_ context.Context, _, id int64) error {
	s.deletedID = id
	return s.deleteErr
}

func TestEnrollmentCreateRejectsInvalidInputBeforeStore(t *testing.T) {
	store := &enrollmentTestStore{}
	_, validation, err := NewEnrollmentService(store).Create(context.Background(), 10, model.NewEnrollment{CourseID: 0, TahunAkademik: "2025-2026-Ganjil"})
	if err != nil || validation == nil || store.created.CourseID != 0 {
		t.Fatalf("validation=%v err=%v store=%+v", validation, err, store.created)
	}
}

func TestEnrollmentCreateMapsMissingCourseToValidation(t *testing.T) {
	store := &enrollmentTestStore{createErr: repository.ErrNotFound}
	_, validation, err := NewEnrollmentService(store).Create(context.Background(), 10, model.NewEnrollment{CourseID: 2, TahunAkademik: "2025/2026-Ganjil"})
	if err != nil || validation == nil || len(validation["course_id"]) == 0 {
		t.Fatalf("validation=%v err=%v", validation, err)
	}
}

func TestEnrollmentDeleteMapsNotFoundAndForbidden(t *testing.T) {
	for _, test := range []struct {
		name  string
		cause error
		want  error
	}{{"not found", repository.ErrNotFound, ErrEnrollmentNotFound}, {"forbidden", repository.ErrForbidden, ErrEnrollmentForbidden}} {
		t.Run(test.name, func(t *testing.T) {
			store := &enrollmentTestStore{deleteErr: test.cause}
			err := NewEnrollmentService(store).Delete(context.Background(), 10, 7)
			if !errors.Is(err, test.want) {
				t.Fatalf("got %v want %v", err, test.want)
			}
		})
	}
}
