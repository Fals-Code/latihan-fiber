package service

import (
	"strings"

	"tugas1-go/pertemuan-5-authentication-security/app/model"
)

func ValidateAchievementCreate(req model.CreateAchievementRequest, fields map[string]bool) (model.CreateAchievementRequest, map[string]string) {
	errs := map[string]string{}
	if !fields["name"] {
		errs["name"] = "wajib diisi"
	} else {
		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" {
			errs["name"] = "wajib diisi"
		}
	}
	if !fields["student_id"] {
		errs["student_id"] = "wajib diisi"
	} else if req.StudentID <= 0 {
		errs["student_id"] = "harus berupa angka positif"
	}
	if !fields["rank"] {
		errs["rank"] = "wajib diisi"
	} else if req.Rank <= 0 {
		errs["rank"] = "harus berupa angka positif"
	}
	return req, errs
}

func ValidateAchievementReplace(req model.ReplaceAchievementRequest, fields map[string]bool) (model.ReplaceAchievementRequest, map[string]string) {
	errs := map[string]string{}
	if !fields["name"] {
		errs["name"] = "wajib diisi pada PUT"
	} else {
		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" {
			errs["name"] = "tidak boleh kosong"
		}
	}
	if !fields["student_id"] {
		errs["student_id"] = "wajib diisi pada PUT"
	} else if req.StudentID <= 0 {
		errs["student_id"] = "harus berupa angka positif"
	}
	if !fields["rank"] {
		errs["rank"] = "wajib diisi pada PUT"
	} else if req.Rank <= 0 {
		errs["rank"] = "harus berupa angka positif"
	}
	return req, errs
}

func ApplyAchievementPatch(current model.Achievement, req model.PatchAchievementRequest) (model.Achievement, map[string]string) {
	errs := map[string]string{}
	if req.Name != nil {
		current.Name = strings.TrimSpace(*req.Name)
		if current.Name == "" {
			errs["name"] = "tidak boleh kosong"
		}
	}
	if req.StudentID != nil {
		if *req.StudentID <= 0 {
			errs["student_id"] = "harus berupa angka positif"
		} else {
			current.StudentID = *req.StudentID
		}
	}
	if req.Rank != nil {
		if *req.Rank <= 0 {
			errs["rank"] = "harus berupa angka positif"
		} else {
			current.Rank = *req.Rank
		}
	}
	return current, errs
}

func IsEmptyAchievementPatch(req model.PatchAchievementRequest) bool {
	return req.Name == nil && req.StudentID == nil && req.Rank == nil
}
