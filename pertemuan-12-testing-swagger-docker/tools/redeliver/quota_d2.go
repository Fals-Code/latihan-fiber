//go:build d2

package main

import (
	"encoding/json"
	"fmt"

	"tugas1-go/pertemuan-12-testing-swagger-docker/app/model"
)

func quotaRecipientJobs(job model.NotificationJob) (bool, []model.NotificationJob, error) {
	if job.EventType != model.EventCourseQuotaChanged {
		return false, nil, nil
	}
	var event model.QuotaChangedEvent
	if err := json.Unmarshal(job.Payload, &event); err != nil {
		return true, nil, fmt.Errorf("membaca event kuota: %w", err)
	}
	if event.CourseID < 1 || event.Quota < 1 {
		return true, nil, fmt.Errorf("event kuota tidak valid")
	}
	jobs := make([]model.NotificationJob, 0, len(event.UserIDs))
	for _, userID := range event.UserIDs {
		if userID < 1 {
			return true, nil, fmt.Errorf("penerima kuota tidak valid")
		}
		recipientEvent := event
		recipientEvent.UserIDs = []int{userID}
		payload, err := json.Marshal(recipientEvent)
		if err != nil {
			return true, nil, fmt.Errorf("menyusun event penerima: %w", err)
		}
		// OutboxID asli dipertahankan agar salinan dengan TaskID acak tetap
		// menguji claim yang sama, bukan membuat notifikasi baru.
		recipientJob := job
		recipientJob.Payload = payload
		jobs = append(jobs, recipientJob)
	}
	return true, jobs, nil
}
