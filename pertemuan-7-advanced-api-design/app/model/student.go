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
	NIM   int     `json:"nim" validate:"required,gt=0"`
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
	NIM      *int     `json:"nim,omitempty" validate:"omitempty,gt=0"`
	Name     *string  `json:"name,omitempty" validate:"omitempty,min=1"`
	Grade    *float64 `json:"grade,omitempty" validate:"omitempty,gte=0,lte=100"`
	IsActive *bool    `json:"is_active,omitempty"`
}

type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    *Meta  `json:"meta,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}
type Meta struct {
	Page       int    `json:"page,omitempty"`
	Limit      int    `json:"limit,omitempty"`
	Total      int    `json:"total,omitempty"`
	TotalPages int    `json:"total_pages,omitempty"`
	NextCursor string `json:"next_cursor,omitempty"`
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
