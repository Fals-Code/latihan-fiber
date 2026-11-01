package model

import "time"

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

type AssignRoleRequest struct {
	Role string `json:"role"`
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type UpdateUserRequest struct {
	Username *string `json:"username,omitempty" validate:"omitnil,min=3"`
	Email    *string `json:"email,omitempty" validate:"omitnil,email"`
	IsActive *bool   `json:"is_active,omitempty" validate:"omitnil"`
}

type UserListQuery struct {
	Limit  int
	Cursor string
}
