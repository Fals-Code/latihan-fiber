package service

import (
	"strings"

	"tugas1-go/pertemuan-5-authentication-security/app/model"
)

// ValidateCreate memeriksa business rules untuk POST student.
func ValidateCreate(
	req model.CreateStudentRequest,
	fields map[string]bool,
) (model.CreateStudentRequest, map[string]string) {
	errs := map[string]string{}

	if !fields["nim"] {
		errs["nim"] = "wajib diisi"
	} else if req.NIM <= 0 {
		errs["nim"] = "harus berupa angka positif"
	}

	if !fields["name"] {
		errs["name"] = "wajib diisi"
	} else {
		req.Name = strings.TrimSpace(req.Name)

		if req.Name == "" {
			errs["name"] = "wajib diisi"
		}
	}

	if !fields["grade"] {
		errs["grade"] = "wajib diisi"
	} else if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "harus berada antara 0 sampai 100"
	}

	return req, errs
}

// ValidateReplace memeriksa business rules untuk PUT.
// Semua field wajib dikirim karena PUT mengganti seluruh data student.
func ValidateReplace(
	req model.ReplaceStudentRequest,
	fields map[string]bool,
) (model.ReplaceStudentRequest, map[string]string) {
	errs := map[string]string{}

	if !fields["nim"] {
		errs["nim"] = "wajib diisi pada PUT"
	} else if req.NIM <= 0 {
		errs["nim"] = "harus berupa angka positif"
	}

	if !fields["name"] {
		errs["name"] = "wajib diisi pada PUT"
	} else {
		req.Name = strings.TrimSpace(req.Name)

		if req.Name == "" {
			errs["name"] = "tidak boleh kosong"
		}
	}

	if !fields["grade"] {
		errs["grade"] = "wajib diisi pada PUT"
	} else if req.Grade < 0 || req.Grade > 100 {
		errs["grade"] = "harus berada antara 0 sampai 100"
	}

	if !fields["is_active"] {
		errs["is_active"] = "wajib diisi pada PUT"
	}

	return req, errs
}

// ApplyPatch menerapkan field PATCH yang dikirim ke data student lama.
// Field nil berarti tidak diubah.
func ApplyPatch(
	current model.Student,
	req model.PatchStudentRequest,
) (model.Student, map[string]string) {
	errs := map[string]string{}

	if req.NIM != nil {
		if *req.NIM <= 0 {
			errs["nim"] = "harus berupa angka positif"
		} else {
			current.NIM = *req.NIM
		}
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)

		if name == "" {
			errs["name"] = "tidak boleh kosong"
		} else {
			current.Name = name
		}
	}

	if req.Grade != nil {
		if *req.Grade < 0 || *req.Grade > 100 {
			errs["grade"] = "harus berada antara 0 sampai 100"
		} else {
			current.Grade = *req.Grade
		}
	}

	if req.IsActive != nil {
		current.IsActive = *req.IsActive
	}

	return current, errs
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
