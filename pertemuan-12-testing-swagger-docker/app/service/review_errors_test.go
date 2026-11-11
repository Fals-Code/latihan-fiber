package service

import (
	"errors"
	"fmt"
	"testing"

	"tugas1-go/pertemuan-12-testing-swagger-docker/app/repository"
	"tugas1-go/pertemuan-12-testing-swagger-docker/helper"
)

func TestTranslateReviewError(t *testing.T) {
	tests := []struct {
		err        error
		wantStatus int
		wantCode   string
	}{
		{repository.ErrInvalidID, 400, helper.CodeBadRequest},
		{repository.ErrNotFound, 404, helper.CodeNotFound},
		{repository.ErrDuplicate, 409, "ALREADY_REVIEWED"},
		{repository.ErrNotOwner, 403, helper.CodeForbidden},
		{repository.ErrOwnReview, 409, "OWN_REVIEW"},
		{repository.ErrAlreadyVoted, 409, "ALREADY_VOTED"},
		{repository.ErrReplyLimit, 409, "REPLY_LIMIT"},
		{fmt.Errorf("outer: %w", repository.ErrDuplicate), 409, "ALREADY_REVIEWED"},
		{errors.New("MongoDB connection lost"), 500, helper.CodeInternal},
	}
	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			var appErr *helper.AppError
			if !errors.As(translateReviewError(tt.err), &appErr) {
				t.Fatal("result is not *helper.AppError")
			}
			if appErr.Status != tt.wantStatus || appErr.Code != tt.wantCode {
				t.Errorf("got %d %s, want %d %s", appErr.Status, appErr.Code, tt.wantStatus, tt.wantCode)
			}
		})
	}
}
