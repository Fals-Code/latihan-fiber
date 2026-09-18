package service

import (
	"testing"

	"tugas1-go/pertemuan-6-authorization-rbac/app/model"
	"tugas1-go/pertemuan-6-authorization-rbac/helper"
)

func TestCanAccessStudent(t *testing.T) {
	perms := helper.NewPermissionSet(map[string][]string{
		"admin": {"student:read:any", "student:update:any"},
		"staff": {"student:read:any"},
		"user":  {},
	})
	cases := []struct {
		name       string
		user       model.AuthUser
		owner      int
		permission string
		want       bool
	}{
		{"user owner", model.AuthUser{UserID: 1, Role: "user"}, 1, "student:read:any", true},
		{"staff owner", model.AuthUser{UserID: 1, Role: "staff"}, 1, "student:update:any", true},
		{"admin owner", model.AuthUser{UserID: 1, Role: "admin"}, 1, "student:update:any", true},
		{"admin any", model.AuthUser{UserID: 1, Role: "admin"}, 2, "student:read:any", true},
		{"staff any", model.AuthUser{UserID: 1, Role: "staff"}, 2, "student:read:any", true},
		{"user denied", model.AuthUser{UserID: 1, Role: "user"}, 2, "student:read:any", false},
		{"staff update denied", model.AuthUser{UserID: 1, Role: "staff"}, 2, "student:update:any", false},
		{"admin update any", model.AuthUser{UserID: 1, Role: "admin"}, 2, "student:update:any", true},
		{"nil denied", model.AuthUser{UserID: 1, Role: "user"}, 2, "student:read:any", false},
		{"unknown role", model.AuthUser{UserID: 1, Role: "unknown"}, 2, "student:read:any", false},
		{"unknown permission", model.AuthUser{UserID: 1, Role: "admin"}, 2, "student:unknown", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			set := perms
			if tc.name == "nil denied" {
				set = nil
			}
			if got := CanAccessStudent(tc.user, tc.owner, set, tc.permission); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLegacyNilOwnerCannotMatchCaller(t *testing.T) {
	if canAccessStudentOwner(model.AuthUser{UserID: 1, Role: "user"}, nil, helper.NewPermissionSet(map[string][]string{"user": {}}), "student:read:any") {
		t.Fatal("legacy nil owner must not become caller-owned")
	}
}
