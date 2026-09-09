package service

import (
	"testing"

	"tugas1-go/pertemuan-5-authentication-security/app/model"
)

func achievementFields() map[string]bool {
	return map[string]bool{"name": true, "student_id": true, "rank": true}
}

func TestValidateAchievementCreateValid(t *testing.T) {
	result, errs := ValidateAchievementCreate(model.CreateAchievementRequest{Name: "  Juara  ", StudentID: 1, Rank: 2}, achievementFields())
	if len(errs) != 0 {
		t.Fatalf("expected no validation errors, got %v", errs)
	}
	if result.Name != "Juara" {
		t.Errorf("expected trimmed name, got %q", result.Name)
	}
}

func TestValidateAchievementCreateRejectsInvalidValues(t *testing.T) {
	_, errs := ValidateAchievementCreate(model.CreateAchievementRequest{Name: " ", StudentID: 0, Rank: 0}, achievementFields())
	for _, field := range []string{"name", "student_id", "rank"} {
		if _, ok := errs[field]; !ok {
			t.Errorf("expected validation error for %s", field)
		}
	}
}

func TestApplyAchievementPatchOnlyChangesProvidedFields(t *testing.T) {
	name := "  Updated  "
	rank := 3
	current := model.Achievement{ID: 1, Name: "Old", StudentID: 2, Rank: 1}
	result, errs := ApplyAchievementPatch(current, model.PatchAchievementRequest{Name: &name, Rank: &rank})
	if len(errs) != 0 {
		t.Fatalf("expected no validation errors, got %v", errs)
	}
	if result.Name != "Updated" || result.Rank != 3 || result.StudentID != current.StudentID {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestIsEmptyAchievementPatch(t *testing.T) {
	if !IsEmptyAchievementPatch(model.PatchAchievementRequest{}) {
		t.Fatal("expected empty patch")
	}
	name := "x"
	if IsEmptyAchievementPatch(model.PatchAchievementRequest{Name: &name}) {
		t.Fatal("expected non-empty patch")
	}
}
