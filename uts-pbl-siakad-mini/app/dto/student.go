package dto

type StudentListQuery struct {
	Page     string `query:"page"`
	PerPage  string `query:"per_page"`
	Prodi    string `query:"prodi"`
	Angkatan string `query:"angkatan"`
	Search   string `query:"search"`
	Sort     string `query:"sort"`
}

type CreateStudentRequest struct {
	NIM         string   `json:"nim"`
	Nama        string   `json:"nama"`
	Email       string   `json:"email"`
	Prodi       string   `json:"prodi"`
	Angkatan    *int     `json:"angkatan"`
	IPKTerkahir *float64 `json:"ipk_terakhir"`
}

type StudentResponse struct {
	ID          int64   `json:"id"`
	NIM         string  `json:"nim"`
	Nama        string  `json:"nama"`
	Prodi       string  `json:"prodi"`
	Angkatan    int     `json:"angkatan"`
	IPKTerakhir float64 `json:"ipk_terakhir"`
}

type PaginationMeta struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	Total       int64 `json:"total"`
	LastPage    int   `json:"last_page"`
}

type StudentListEnvelope struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Data    []StudentResponse `json:"data"`
	Meta    PaginationMeta    `json:"meta"`
}
