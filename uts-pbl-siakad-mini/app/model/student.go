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
