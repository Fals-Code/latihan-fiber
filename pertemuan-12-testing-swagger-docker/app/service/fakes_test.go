package service_test

import (
	"context"
	"strconv"
	"sync"
	"time"

	"tugas1-go/pertemuan-12-testing-swagger-docker/app/model"
	"tugas1-go/pertemuan-12-testing-swagger-docker/app/repository"
)

var (
	_ repository.ReviewRepository = (*fakeReviewRepo)(nil)
	_ repository.CourseRepository = (*fakeCourseRepo)(nil)
)

type fakeReviewRepo struct {
	repository.ReviewRepository
	mu      sync.Mutex
	reviews []model.Review
}

func (f *fakeReviewRepo) Create(_ context.Context, review model.Review) (model.Review, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, existing := range f.reviews {
		if existing.CourseID == review.CourseID && existing.UserID == review.UserID {
			return model.Review{}, repository.ErrDuplicate
		}
	}
	review.ID = strconv.Itoa(len(f.reviews) + 1)
	review.CreatedAt = time.Now()
	review.UpdatedAt = review.CreatedAt
	if review.Tags == nil {
		review.Tags = []string{}
	}
	review.Replies = []model.Reply{}
	f.reviews = append(f.reviews, review)
	return review, nil
}

func (f *fakeReviewRepo) FindByCourse(_ context.Context, query model.ReviewQuery) ([]model.Review, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	result := []model.Review{}
	for i := len(f.reviews) - 1; i >= 0; i-- {
		review := f.reviews[i]
		if review.CourseID == query.CourseID {
			result = append(result, review)
		}
	}
	return result, nil
}

type fakeCourseRepo struct {
	repository.CourseRepository
	courses  map[int]model.Course
	enrolled map[[2]int]bool
}

func (f *fakeCourseRepo) FindByID(_ context.Context, id int) (model.Course, error) {
	course, ok := f.courses[id]
	if !ok {
		return model.Course{}, repository.ErrNotFound
	}
	return course, nil
}

func (f *fakeCourseRepo) IsEnrolled(_ context.Context, courseID, userID int) (bool, error) {
	return f.enrolled[[2]int{courseID, userID}], nil
}
