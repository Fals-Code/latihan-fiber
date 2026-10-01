package service

import (
	"context"
	"strconv"
	"strings"

	"tugas1-go/uts-pbl-siakad-mini/app/model"
)

type CourseStore interface {
	List(context.Context, model.CourseFilters) ([]model.Course, error)
}

type CourseService struct{ store CourseStore }

func NewCourseService(store CourseStore) *CourseService { return &CourseService{store: store} }

func (s *CourseService) List(ctx context.Context, query map[string]string) ([]model.Course, map[string][]string, error) {
	filters := model.CourseFilters{Search: strings.TrimSpace(query["search"])}
	validation := make(map[string][]string)
	if raw := query["semester"]; raw != "" {
		semester, err := strconv.Atoi(raw)
		if err != nil || semester < 1 || semester > 14 {
			validation["semester"] = []string{"Semester harus antara 1 dan 14"}
		} else {
			filters.Semester = &semester
		}
	}
	if raw := query["available"]; raw != "" {
		if raw != "true" && raw != "false" {
			validation["available"] = []string{"Available harus true atau false"}
		} else {
			filters.Available = raw == "true"
		}
	}
	if len(validation) > 0 {
		return nil, validation, nil
	}
	courses, err := s.store.List(ctx, filters)
	return courses, nil, err
}
