package service

import (
	"strings"

	"tugas1-go/pertemuan-6-authorization-rbac/app/model"
	"tugas1-go/pertemuan-6-authorization-rbac/helper"
)

func CanAccessUser(current model.AuthUser, targetID int, perms *helper.PermissionSet, anyPermission string) bool {
	if current.UserID == targetID {
		return true
	}
	return perms.Can(current.Role, anyPermission)
}

func ValidateAssignRole(current model.AuthUser, targetID int, req model.AssignRoleRequest, perms *helper.PermissionSet) map[string]string {
	errs := map[string]string{}
	req.Role = strings.TrimSpace(req.Role)
	if req.Role == "" {
		errs["role"] = "wajib diisi"
		return errs
	}
	if !perms.IsKnownRole(req.Role) {
		errs["role"] = "role tidak dikenal"
	}
	if current.UserID == targetID {
		errs["role"] = "tidak boleh mengubah role sendiri"
	}
	return errs
}
