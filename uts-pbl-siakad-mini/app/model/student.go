package model

type Student struct {
	ID          int64   `json:"id"`
	NIM         string  `json:"nim"`
	Nama        string  `json:"nama"`
	Prodi       string  `json:"prodi"`
	Angkatan    int     `json:"angkatan"`
	IPKTerakhir float64 `json:"ipk_terakhir"`
}

type StudentFilters struct {
	Page     int
	PerPage  int
	Prodi    string
	Angkatan *int
	Search   string
	Sort     string
}

type NewStudent struct {
	NIM         string
	Nama        string
	Email       string
	Prodi       string
	Angkatan    int
	IPKTerakhir float64
	Password    string
}

type StudentCourse struct {
	EnrollmentID  int64  `json:"enrollment_id"`
	CourseID      int64  `json:"course_id"`
	KodeMK        string `json:"kode_mk"`
	NamaMK        string `json:"nama_mk"`
	SKS           int    `json:"sks"`
	Semester      int    `json:"semester"`
	TahunAkademik string `json:"tahun_akademik"`
}

type StudentDetail struct {
	Student
	UserID   int64           `json:"-"`
	Courses  []StudentCourse `json:"courses"`
	TotalSKS int             `json:"total_sks"`
	BatasSKS int             `json:"batas_sks"`
}

type StudentUpdate struct {
	Nama        *string
	Prodi       *string
	Angkatan    *int
	IPKTerakhir *float64
}
