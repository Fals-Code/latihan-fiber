//go:build d2

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/hibiken/asynq"

	"tugas1-go/pertemuan-12-testing-swagger-docker/app/model"
	"tugas1-go/pertemuan-12-testing-swagger-docker/app/repository"
	"tugas1-go/pertemuan-12-testing-swagger-docker/helper"
)

// Satu task hanya memiliki satu penerima, sehingga kegagalannya tidak
// mengulang pengiriman penerima lain yang sudah berhasil.
func (w *NotificationWorker) handleQuotaEvent(ctx context.Context, job model.NotificationJob) (bool, error) {
	if job.EventType != model.EventCourseQuotaChanged {
		return false, nil
	}
	var event model.QuotaChangedEvent
	if err := json.Unmarshal(job.Payload, &event); err != nil {
		return true, fmt.Errorf("%w: payload kuota rusak: %v", asynq.SkipRetry, err)
	}
	if event.CourseID < 1 || event.Quota < 1 || len(event.UserIDs) != 1 || event.UserIDs[0] < 1 {
		return true, fmt.Errorf("%w: task kuota harus memiliki tepat satu penerima valid", asynq.SkipRetry)
	}
	userID := event.UserIDs[0]
	user, err := w.users.FindByID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return true, fmt.Errorf("%w: user %d dihapus", asynq.SkipRetry, userID)
	}
	if err != nil {
		return true, err
	}
	subject := fmt.Sprintf("KRS: kuota %s %s berubah menjadi %d", event.CourseCode, event.CourseName, event.Quota)
	msg := model.Message{To: user.Email, Subject: subject, Body: fmt.Sprintf("Halo %s, %s.", user.Username, subject)}
	alreadySent := false
	err = w.tx.WithTxOnce(ctx, func(tx repository.DBTX) error {
		// Claim per event/penerima/channel ikut rollback bila pengiriman gagal.
		claimed, err := w.notifications.WithTx(tx).Claim(ctx, model.Notification{OutboxID: job.OutboxID, UserID: user.ID, Channel: "email", Subject: msg.Subject, Body: msg.Body})
		if err != nil {
			return err
		}
		if !claimed {
			alreadySent = true
			return nil
		}
		return w.mailer.Send(ctx, msg)
	})
	if errors.Is(err, helper.ErrPermanentDelivery) {
		return true, fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}
	if err != nil {
		return true, err
	}
	retry, _ := asynq.GetRetryCount(ctx)
	w.logger.Info("notification_processed", slog.Int64("outbox_id", job.OutboxID), slog.Int("user_id", userID), slog.String("event", job.EventType), slog.Int("retry", retry), slog.Bool("already_sent", alreadySent))
	return true, nil
}
