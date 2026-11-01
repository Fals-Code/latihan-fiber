//go:build d2

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/hibiken/asynq"

	"tugas1-go/pertemuan-11-background-job/app/model"
)

var errQuotaQueueUnavailable = errors.New("queue terputus setelah dua penerima")

// Fake ini hanya menguji kontrak enqueue. Ia tidak membuktikan transaksi
// RelayOnce, published_at di PostgreSQL, maupun delivery worker/Redis nyata.
type quotaQueueFake struct {
	t         *testing.T
	accepted  map[string]*asynq.Task
	failOnce  bool
	conflicts int
}

func (q *quotaQueueFake) EnqueueContext(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	q.t.Helper()
	values := map[asynq.OptionType]interface{}{}
	for _, opt := range opts {
		values[opt.Type()] = opt.Value()
	}
	if values[asynq.QueueOpt] != model.QueueNotifications || values[asynq.MaxRetryOpt] != 5 || values[asynq.TimeoutOpt] != 30*time.Second || values[asynq.RetentionOpt] != 24*time.Hour {
		q.t.Fatalf("kebijakan task berubah: %v", values)
	}
	id, ok := values[asynq.TaskIDOpt].(string)
	if !ok || id == "" {
		q.t.Fatal("task harus memiliki TaskID stabil")
	}
	if _, exists := q.accepted[id]; exists {
		q.conflicts++
		return nil, fmt.Errorf("task sudah ada: %w", asynq.ErrTaskIDConflict)
	}
	if q.failOnce && len(q.accepted) == 2 {
		q.failOnce = false
		return nil, errQuotaQueueUnavailable
	}
	q.accepted[id] = task
	return &asynq.TaskInfo{ID: id}, nil
}

func TestQuotaRelayResumesPartialEnqueue(t *testing.T) {
	event := model.QuotaChangedEvent{CourseID: 2, CourseCode: "IF102", CourseName: "Basis Data Lanjut", Quota: 40, UserIDs: []int{11, 12, 13, 14, 15}}
	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	outbox := model.OutboxEvent{ID: 77, EventType: model.EventCourseQuotaChanged, Payload: payload}
	queue := &quotaQueueFake{t: t, accepted: map[string]*asynq.Task{}, failOnce: true}
	relay := &OutboxRelay{queue: queue}
	if err := relay.publish(context.Background(), outbox); !errors.Is(err, errQuotaQueueUnavailable) {
		t.Fatalf("enqueue parsial harus mengembalikan error: %v", err)
	}
	if len(queue.accepted) != 2 {
		t.Fatalf("diterima sebelum putus = %d, ingin 2", len(queue.accepted))
	}
	if err := relay.publish(context.Background(), outbox); err != nil {
		t.Fatal(err)
	}
	if len(queue.accepted) != 5 || queue.conflicts != 2 {
		t.Fatalf("pemulihan: diterima=%d conflict=%d, ingin 5/2", len(queue.accepted), queue.conflicts)
	}
	for _, userID := range event.UserIDs {
		id := fmt.Sprintf("outbox-77-user-%d-email", userID)
		task, exists := queue.accepted[id]
		if !exists {
			t.Fatalf("task %s tidak ada", id)
		}
		if task.Type() != model.TaskSendNotification {
			t.Fatalf("type task salah: %s", task.Type())
		}
		var job model.NotificationJob
		if err := json.Unmarshal(task.Payload(), &job); err != nil {
			t.Fatal(err)
		}
		var recipient model.QuotaChangedEvent
		if err := json.Unmarshal(job.Payload, &recipient); err != nil {
			t.Fatal(err)
		}
		if job.OutboxID != 77 || job.EventType != outbox.EventType || len(recipient.UserIDs) != 1 || recipient.UserIDs[0] != userID || recipient.CourseID != event.CourseID || recipient.CourseCode != event.CourseCode || recipient.CourseName != event.CourseName || recipient.Quota != event.Quota {
			t.Fatalf("snapshot penerima %d tidak sesuai: job=%+v event=%+v", userID, job, recipient)
		}
	}
}

func TestQuotaWorkerRejectsMultipleRecipients(t *testing.T) {
	// Tanpa dependency worker: payload tidak sah harus ditolak sebelum DB/mail.
	worker := &NotificationWorker{}
	payload := json.RawMessage(`{"course_id":2,"quota":40,"user_ids":[11,12]}`)
	handled, err := worker.handleQuotaEvent(context.Background(), model.NotificationJob{EventType: model.EventCourseQuotaChanged, Payload: payload})
	if !handled || !errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("ingin SkipRetry: handled=%v err=%v", handled, err)
	}
}
