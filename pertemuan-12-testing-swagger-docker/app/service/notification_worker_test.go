package service

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/hibiken/asynq"

	"tugas1-go/pertemuan-12-testing-swagger-docker/app/model"
	"tugas1-go/pertemuan-12-testing-swagger-docker/app/repository"
)

type fakeUsers struct {
	repository.UserRepository
	users map[int]model.User
}

func (f *fakeUsers) FindByID(_ context.Context, id int) (model.User, error) {
	user, ok := f.users[id]
	if !ok {
		return model.User{}, repository.ErrNotFound
	}
	return user, nil
}

func newNotificationTask(t *testing.T, eventType string, event any) *asynq.Task {
	t.Helper()
	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}
	job, err := json.Marshal(model.NotificationJob{OutboxID: 1, EventType: eventType, Payload: payload})
	if err != nil {
		t.Fatalf("marshal job: %v", err)
	}
	return asynq.NewTask(model.TaskSendNotification, job)
}

func TestHandleSend_PermanentFailures(t *testing.T) {
	worker := NewNotificationWorker(
		&fakeUsers{users: map[int]model.User{7: {ID: 7, Username: "mhsuji", Email: "mhsuji@student.unair.ac.id"}}},
		nil, nil, nil, slog.New(slog.NewTextHandler(io.Discard, nil)),
	)
	tests := []struct {
		name string
		task *asynq.Task
	}{
		{name: "malformed job payload", task: asynq.NewTask(model.TaskSendNotification, []byte("not json"))},
		{name: "deleted user", task: newNotificationTask(t, model.EventEnrollmentCreated, model.EnrollmentEvent{UserID: 99})},
		{name: "unknown event", task: newNotificationTask(t, "enrollment.unknown", model.EnrollmentEvent{UserID: 7})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := worker.HandleSend(context.Background(), tt.task)
			if !errors.Is(err, asynq.SkipRetry) {
				t.Errorf("error = %v, want error wrapping asynq.SkipRetry", err)
			}
		})
	}
}
