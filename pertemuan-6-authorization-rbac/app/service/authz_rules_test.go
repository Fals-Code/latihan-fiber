package service

import (
	"testing"

	"tugas1-go/pertemuan-6-authorization-rbac/app/model"
	"tugas1-go/pertemuan-6-authorization-rbac/helper"
)

func TestCanAccessUser(t *testing.T) {
	perms := helper.NewPermissionSet(map[string][]string{"admin": {"user:read:any"}, "user": {}})
	if !CanAccessUser(model.AuthUser{UserID: 1, Role: "user"}, 1, perms, "user:read:any") {
		t.Fatal("owner should access")
	}
	if !CanAccessUser(model.AuthUser{UserID: 1, Role: "admin"}, 2, perms, "user:read:any") {
		t.Fatal("permitted role should access")
	}
	if CanAccessUser(model.AuthUser{UserID: 1, Role: "user"}, 2, perms, "user:read:any") {
		t.Fatal("unpermitted role should be denied")
	}
	if CanAccessUser(model.AuthUser{UserID: 1, Role: "user"}, 2, nil, "user:read:any") {
		t.Fatal("nil permissions should deny")
	}
}

func TestValidateAssignRole(t *testing.T) {
	perms := helper.NewPermissionSet(map[string][]string{"admin": {"role:assign"}, "user": {}})
	current := model.AuthUser{UserID: 1, Role: "admin"}
	if errs := ValidateAssignRole(current, 2, model.AssignRoleRequest{Role: " user "}, perms); len(errs) != 0 {
		t.Fatalf("valid role errors: %v", errs)
	}
	if errs := ValidateAssignRole(current, 2, model.AssignRoleRequest{}, perms); errs["role"] == "" {
		t.Fatal("blank role should fail")
	}
	if errs := ValidateAssignRole(current, 2, model.AssignRoleRequest{Role: "unknown"}, perms); errs["role"] == "" {
		t.Fatal("unknown role should fail")
	}
	if errs := ValidateAssignRole(current, 1, model.AssignRoleRequest{Role: "user"}, perms); errs["role"] == "" {
		t.Fatal("self-role change should fail")
	}
	if errs := ValidateAssignRole(current, 2, model.AssignRoleRequest{Role: "user"}, perms); len(errs) != 0 {
		t.Fatalf("user should be valid target role: %v", errs)
	}
}
