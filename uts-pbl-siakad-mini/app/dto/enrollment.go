package dto

import "time"

type CreateEnrollmentRequest struct {
	CourseID      int64  `json:"course_id"`
	TahunAkademik string `json:"tahun_akademik"`
}

type EnrollmentResponse struct {
	ID            int64     `json:"id"`
	CourseID      int64     `json:"course_id"`
	TahunAkademik string    `json:"tahun_akademik"`
	CreatedAt     time.Time `json:"created_at"`
}
