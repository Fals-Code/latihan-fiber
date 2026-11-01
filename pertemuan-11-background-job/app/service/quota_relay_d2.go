//go:build d2

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hibiken/asynq"

	"tugas1-go/pertemuan-11-background-job/app/model"
)

func (r *OutboxRelay) publishQuotaEvent(ctx context.Context, e model.OutboxEvent) (bool, error) {
	if e.EventType != model.EventCourseQuotaChanged {
		return false, nil
	}
	var event model.QuotaChangedEvent
	if err := json.Unmarshal(e.Payload, &event); err != nil {
		return true, fmt.Errorf("membaca event kuota outbox %d: %w", e.ID, err)
	}
	if event.CourseID < 1 || event.Quota < 1 {
		return true, fmt.Errorf("event kuota outbox %d tidak valid", e.ID)
	}
	for _, userID := range event.UserIDs {
		if userID < 1 {
			return true, fmt.Errorf("penerima outbox %d tidak valid", e.ID)
		}
	}
	for _, userID := range event.UserIDs {
		recipientEvent := event
		recipientEvent.UserIDs = []int{userID}
		eventPayload, err := json.Marshal(recipientEvent)
		if err != nil {
			return true, fmt.Errorf("menyusun event kuota outbox %d: %w", e.ID, err)
		}
		payload, err := json.Marshal(model.NotificationJob{
			OutboxID: e.ID, EventType: e.EventType, Payload: eventPayload,
		})
		if err != nil {
			return true, fmt.Errorf("menyusun task kuota outbox %d: %w", e.ID, err)
		}
		// ID stabil per penerima membuat pengulangan setelah enqueue parsial
		// melewati task yang sudah ada, lalu melanjutkan penerima berikutnya.
		_, err = r.queue.EnqueueContext(ctx, asynq.NewTask(model.TaskSendNotification, payload),
			asynq.TaskID(fmt.Sprintf("outbox-%d-user-%d-email", e.ID, userID)),
			asynq.Queue(model.QueueNotifications), asynq.MaxRetry(5),
			asynq.Timeout(30*time.Second), asynq.Retention(24*time.Hour))
		if errors.Is(err, asynq.ErrTaskIDConflict) {
			continue
		}
		if err != nil {
			// RelayOnce hanya menandai published setelah SEMUA penerima masuk
			// queue. Event parsial tetap pending untuk putaran berikutnya.
			return true, fmt.Errorf("enqueue outbox %d penerima %d: %w", e.ID, userID, err)
		}
	}
	// Kelas tanpa peserta tetap sah; tidak ada pengiriman yang perlu ditunggu.
	return true, nil
}
