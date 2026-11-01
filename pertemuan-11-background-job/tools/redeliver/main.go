// Alat buatan sendiri berdasarkan C.5 Modul 11; belum diverifikasi.
// Mengirim salinan dengan TaskID unik dan tidak mengubah published_at.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"os"
	"time"
	"tugas1-go/pertemuan-11-background-job/app/model"
	"tugas1-go/pertemuan-11-background-job/config"
	"tugas1-go/pertemuan-11-background-job/database"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	outboxID := flag.Int64("outbox", 0, "ID event yang akan disalin")
	copies := flag.Int("copies", 5, "jumlah salinan (1-100)")
	flag.Parse()
	if *outboxID < 1 || *copies < 1 || *copies > 100 {
		return fmt.Errorf("gunakan -outbox ID positif dan -copies 1..100")
	}
	config.LoadEnv()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := database.NewPool(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()
	var job model.NotificationJob
	job.OutboxID = *outboxID
	if err := pool.QueryRow(ctx, "SELECT event_type,payload FROM outbox WHERE id=$1", *outboxID).Scan(&job.EventType, &job.Payload); err != nil {
		return fmt.Errorf("membaca event: %w", err)
	}
	handled, jobs, err := quotaRecipientJobs(job)
	if err != nil {
		return err
	}
	if !handled {
		jobs = []model.NotificationJob{job}
	}
	client := asynq.NewClient(config.RedisOpt())
	defer client.Close()
	for i := 0; i < *copies; i++ {
		for _, recipientJob := range jobs {
			payload, err := json.Marshal(recipientJob)
			if err != nil {
				return err
			}
			task := asynq.NewTask(model.TaskSendNotification, payload)
			id := fmt.Sprintf("redeliver-%d-%s", *outboxID, uuid.NewString())
			if _, err := client.EnqueueContext(ctx, task, asynq.TaskID(id), asynq.Queue(model.QueueNotifications), asynq.MaxRetry(5), asynq.Timeout(30*time.Second), asynq.Retention(24*time.Hour)); err != nil {
				return fmt.Errorf("salinan %d/%d: %w", i+1, *copies, err)
			}
			fmt.Printf("enqueued outbox_id=%d copy=%d task_id=%s\n", *outboxID, i+1, id)
		}
	}
	return nil
}
