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
	Name      string `json:"name"`
	StudentID int    `json:"student_id"`
	Rank      int    `json:"rank"`
}

type ReplaceAchievementRequest struct {
	Name      string `json:"name"`
	StudentID int    `json:"student_id"`
	Rank      int    `json:"rank"`
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
}

func (q ListQueryAchievement) Offset() int {
	return (q.Page - 1) * q.Limit
}
