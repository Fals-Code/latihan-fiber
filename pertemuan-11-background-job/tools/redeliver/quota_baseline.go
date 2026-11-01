//go:build !d2

package main

import "tugas1-go/pertemuan-11-background-job/app/model"

func quotaRecipientJobs(job model.NotificationJob) (bool, []model.NotificationJob, error) {
	return false, nil, nil
}
