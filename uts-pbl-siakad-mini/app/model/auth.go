package model

type User struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"`
	Role     string `json:"role"`
}

type StudentProfile struct {
	NIM      string `json:"nim"`
	Nama     string `json:"nama"`
	Prodi    string `json:"prodi"`
	Angkatan int    `json:"angkatan"`
}

type AuthenticatedUser struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
}

type CurrentUser struct {
	ID      int64           `json:"id"`
	Email   string          `json:"email"`
	Role    string          `json:"role"`
	Student *StudentProfile `json:"student,omitempty"`
}
