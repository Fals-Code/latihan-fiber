package model

import "time"

type Student struct {
	ID        int       `json:"id"`
	NIM       int       `json:"nim"`
	Name      string    `json:"name"`
	Grade     float64   `json:"grade"`
	IsActive  bool      `json:"is_active"`
	OwnerID   *int      `json:"owner_id,omitempty"`
	CreatedAt time.Time `json:"-"`
}

type CreateStudentRequest struct {
	NIM   int     `json:"nim" validate:"required,studentnim"`
	Name  string  `json:"name" validate:"required,min=1"`
	Grade float64 `json:"grade" validate:"gte=0,lte=100"`
}
type ReplaceStudentRequest struct {
	NIM      int     `json:"nim" validate:"required,gt=0"`
	Name     string  `json:"name" validate:"required,min=1"`
	Grade    float64 `json:"grade" validate:"gte=0,lte=100"`
	IsActive bool    `json:"is_active"`
}
type PatchStudentRequest struct {
	NIM      *int     `json:"nim" validate:"omitnil,gt=0"`
	Name     *string  `json:"name" validate:"omitnil,min=3"`
	Grade    *float64 `json:"grade" validate:"omitnil,gte=0,lte=100"`
	IsActive *bool    `json:"is_active" validate:"omitnil"`
}

type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
}
type FailureResponse struct {
	Success   bool              `json:"success"`
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
	RequestID string            `json:"request_id,omitempty"`
}
type Meta struct {
	Limit      int    `json:"limit,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}
type ListQuery struct {
	Page     int
	Limit    int
	Search   string
	Sort     string
	Order    string
	IsActive *bool
	Cursor   string
}

func (q ListQuery) Offset() int { return (q.Page - 1) * q.Limit }
