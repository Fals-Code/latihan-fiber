//go:build d2

package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"tugas1-go/pertemuan-12-testing-swagger-docker/app/model"
)

// claimNotification is shared by KRS and quota notifications under migration 009.
func claimNotification(ctx context.Context, db DBTX, n model.Notification) (bool, error) {
	var id int64
	err := db.QueryRow(ctx, `INSERT INTO notifications (outbox_id,user_id,channel,subject,body)
 VALUES ($1,$2,$3,$4,$5) ON CONFLICT (outbox_id,user_id,channel) DO NOTHING RETURNING id`,
		n.OutboxID, n.UserID, n.Channel, n.Subject, n.Body).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("mencatat notifikasi: %w", err)
	}
	return true, nil
}
