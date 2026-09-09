package service

import (
	"testing"

	"tugas1-go/pertemuan-5-authentication-security/app/model"
)

func TestValidateCreateValid(t *testing.T) {
	req := model.CreateStudentRequest{
		NIM:   20260001,
		Name:  "  Ayuni  ",
		Grade: 0,
	}

	fields := map[string]bool{
		"nim":   true,
		"name":  true,
		"grade": true,
	}

	result, errs := ValidateCreate(req, fields)

	if len(errs) != 0 {
		t.Fatalf("expected no validation errors, got %v", errs)
	}

	if result.Name != "Ayuni" {
		t.Errorf("expected name Ayuni, got %q", result.Name)
	}

	if result.Grade != 0 {
		t.Errorf("expected grade 0, got %v", result.Grade)
	}
}

func TestValidateCreateRejectsInvalidGrade(t *testing.T) {
	req := model.CreateStudentRequest{
		NIM:   20260001,
		Name:  "Ayuni",
		Grade: 101,
	}

	fields := map[string]bool{
		"nim":   true,
		"name":  true,
		"grade": true,
	}

	_, errs := ValidateCreate(req, fields)

	if _, exists := errs["grade"]; !exists {
		t.Fatal("expected validation error for grade")
	}
}

func TestValidateReplaceRequiresIsActive(t *testing.T) {
	req := model.ReplaceStudentRequest{
		NIM:   20260001,
		Name:  "Ayuni",
		Grade: 90,
	}

	fields := map[string]bool{
		"nim":   true,
		"name":  true,
		"grade": true,
	}

	_, errs := ValidateReplace(req, fields)

	if _, exists := errs["is_active"]; !exists {
		t.Fatal("expected validation error for missing is_active")
	}
}

func TestApplyPatchOnlyChangesProvidedFields(t *testing.T) {
	current := model.Student{
		ID:       1,
		NIM:      20260001,
		Name:     "Ayuni",
		Grade:    80,
		IsActive: true,
	}

	newName := "  Falah  "

	req := model.PatchStudentRequest{
		Name: &newName,
	}

	result, errs := ApplyPatch(current, req)

	if len(errs) != 0 {
		t.Fatalf("expected no validation errors, got %v", errs)
	}

	if result.Name != "Falah" {
		t.Errorf("expected name Falah, got %q", result.Name)
	}

	if result.NIM != current.NIM {
		t.Errorf("NIM should not change")
	}

	if result.Grade != current.Grade {
		t.Errorf("grade should not change")
	}

	if result.IsActive != current.IsActive {
		t.Errorf("is_active should not change")
	}
}

func TestCountTotalPages(t *testing.T) {
	result := CountTotalPages(35, 10)

	if result != 4 {
		t.Errorf("expected 4 pages, got %d", result)
	}
}
