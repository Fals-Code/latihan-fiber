package service

import (
	"testing"

	"tugas1-go/pertemuan-7-advanced-api-design/app/model"
	"tugas1-go/pertemuan-7-advanced-api-design/helper"
)

func TestStudentPatchValidationDistinguishesMissingAndEmpty(t *testing.T) {
	var empty model.PatchStudentRequest
	if !IsEmptyPatch(empty) {
		t.Fatal("empty patch should be detected")
	}
	name := ""
	invalid := model.PatchStudentRequest{Name: &name}
	if IsEmptyPatch(invalid) {
		t.Fatal("sent empty name is not an empty patch")
	}
	if errs := helper.ValidateRequest(invalid); errs["name"] == "" {
		t.Fatalf("expected name validation error: %#v", errs)
	}
}

func TestApplyPatchOnlyAppliesPresentFields(t *testing.T) {
	current := model.Student{NIM: 1, Name: "Old", Grade: 70, IsActive: true}
	name := "New"
	updated, errs := ApplyPatch(current, model.PatchStudentRequest{Name: &name})
	if len(errs) != 0 || updated.Name != "New" || updated.NIM != 1 || updated.Grade != 70 || !updated.IsActive {
		t.Fatalf("unexpected patch result: %#v %#v", updated, errs)
	}
}
