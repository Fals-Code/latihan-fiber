//go:build d2

package model

const EventCourseQuotaChanged = "course.quota_changed"

// Recipients are captured when the quota changes, not when a retry happens.
type QuotaChangedEvent struct {
	CourseID   int    `json:"course_id"`
	CourseCode string `json:"course_code"`
	CourseName string `json:"course_name"`
	Quota      int    `json:"quota"`
	UserIDs    []int  `json:"user_ids"`
}
