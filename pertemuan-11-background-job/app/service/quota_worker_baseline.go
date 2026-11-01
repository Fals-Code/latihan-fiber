//go:build !d2

package service

import (
	"context"
	"tugas1-go/pertemuan-11-background-job/app/model"
)

func (w *NotificationWorker) handleQuotaEvent(ctx context.Context, job model.NotificationJob) (bool, error) {
	return false, nil
}
