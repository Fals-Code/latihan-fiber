package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"tugas1-go/uts-pbl-siakad-mini/app/model"
	"tugas1-go/uts-pbl-siakad-mini/app/repository"
)

var enrollmentAcademicYearPattern = regexp.MustCompile(`^[0-9]{4}/[0-9]{4}-(Ganjil|Genap)$`)
var ErrEnrollmentValidation = errors.New("enrollment validation failed")
var ErrEnrollmentNotFound = errors.New("enrollment not found")
var ErrEnrollmentForbidden = errors.New("enrollment access forbidden")

type EnrollmentStore interface {
	Create(context.Context, int64, model.NewEnrollment) (model.Enrollment, error)
	Delete(context.Context, int64, int64) error
}

type EnrollmentService struct{ store EnrollmentStore }

func NewEnrollmentService(store EnrollmentStore) *EnrollmentService {
	return &EnrollmentService{store: store}
}

func (s *EnrollmentService) Create(ctx context.Context, userID int64, input model.NewEnrollment) (model.Enrollment, map[string][]string, error) {
	validation := make(map[string][]string)
	if input.CourseID < 1 {
		validation["course_id"] = []string{"Course ID harus berupa bilangan bulat positif"}
	}
	if !enrollmentAcademicYearPattern.MatchString(input.TahunAkademik) {
		validation["tahun_akademik"] = []string{"Tahun akademik tidak valid"}
	}
	if len(validation) > 0 {
		return model.Enrollment{}, validation, nil
	}
	enrollment, err := s.store.Create(ctx, userID, input)
	if errors.Is(err, repository.ErrInvalidAcademicYear) {
		return model.Enrollment{}, map[string][]string{"tahun_akademik": {"Tahun akademik tidak valid"}}, nil
	}
	if errors.Is(err, repository.ErrNotFound) {
		return model.Enrollment{}, map[string][]string{"course_id": {"Mata kuliah tidak ditemukan"}}, nil
	}
	if errors.Is(err, repository.ErrForbidden) {
		return model.Enrollment{}, nil, ErrEnrollmentForbidden
	}
	return enrollment, nil, err
}

func (s *EnrollmentService) Delete(ctx context.Context, userID, enrollmentID int64) error {
	if enrollmentID < 1 {
		return fmt.Errorf("invalid enrollment id")
	}
	err := s.store.Delete(ctx, userID, enrollmentID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrEnrollmentNotFound
	}
	if errors.Is(err, repository.ErrForbidden) {
		return ErrEnrollmentForbidden
	}
	return err
}

func EnrollmentBusinessError(err error) (int, map[string][]string, bool) {
	switch {
	case errors.Is(err, repository.ErrDuplicateEnrollment):
		return 409, nil, true
	case errors.Is(err, repository.ErrEnrollmentQuotaFull):
		return 422, map[string][]string{"course_id": {"Kuota mata kuliah sudah penuh"}}, true
	case errors.Is(err, repository.ErrEnrollmentSKSLimit):
		message := strings.TrimSpace(strings.TrimPrefix(err.Error(), repository.ErrEnrollmentSKSLimit.Error()+":"))
		return 422, map[string][]string{"course_id": {message}}, true
	default:
		return 0, nil, false
	}
}
