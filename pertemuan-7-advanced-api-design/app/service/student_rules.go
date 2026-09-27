package service

import (
	"strings"

	"tugas1-go/pertemuan-7-advanced-api-design/app/model"
)

// ApplyPatch menerapkan field PATCH yang dikirim ke data student lama.
// Field nil berarti tidak diubah.
func ApplyPatch(current model.Student, req model.PatchStudentRequest) (model.Student, map[string]string) {
	if req.NIM != nil {
		current.NIM = *req.NIM
	}
	if req.Name != nil {
		current.Name = strings.TrimSpace(*req.Name)
	}
	if req.Grade != nil {
		current.Grade = *req.Grade
	}
	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}
	return current, nil
}

// IsEmptyPatch mengecek apakah PATCH tidak membawa satu pun perubahan.
func IsEmptyPatch(req model.PatchStudentRequest) bool {
	return req.NIM == nil &&
		req.Name == nil &&
		req.Grade == nil &&
		req.IsActive == nil
}

// CountTotalPages menghitung jumlah halaman pagination.
func CountTotalPages(total, limit int) int {
	if limit <= 0 {
		return 0
	}

	return (total + limit - 1) / limit
}
