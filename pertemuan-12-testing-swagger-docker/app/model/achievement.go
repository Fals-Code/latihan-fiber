package model

import "time"

type Achievement struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	StudentID int       `json:"student_id"`
	Rank      int       `json:"rank"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateAchievementRequest struct {
	Name      string `json:"name" validate:"required,min=1"`
	StudentID int    `json:"student_id" validate:"required,gt=0"`
	Rank      int    `json:"rank" validate:"required,gt=0"`
}

type ReplaceAchievementRequest struct {
	Name      string `json:"name" validate:"required,min=1"`
	StudentID int    `json:"student_id" validate:"required,gt=0"`
	Rank      int    `json:"rank" validate:"required,gt=0"`
}

type PatchAchievementRequest struct {
	Name      *string `json:"name,omitempty"`
	StudentID *int    `json:"student_id,omitempty"`
	Rank      *int    `json:"rank,omitempty"`
}

type ListQueryAchievement struct {
	Page   int
	Limit  int
	Search string
	Sort   string
	Order  string
	Cursor string
}

type AchievementPage struct {
	Items      []Achievement
	Total      int
	NextCursor string
}

func (q ListQueryAchievement) Offset() int {
	return (q.Page - 1) * q.Limit
}
