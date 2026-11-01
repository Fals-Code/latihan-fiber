//go:build !d2

package service

import (
	"context"

	"tugas1-go/pertemuan-11-background-job/app/model"
)

func (r *OutboxRelay) publishQuotaEvent(ctx context.Context, e model.OutboxEvent) (bool, error) {
	return false, nil
}
