package service

import (
	"tugas1-go/pertemuan-6-authorization-rbac/app/model"
	"tugas1-go/pertemuan-6-authorization-rbac/helper"
)

func CanAccessStudent(current model.AuthUser, ownerID int, perms *helper.PermissionSet, anyPermission string) bool {
	if current.UserID == ownerID {
		return true
	}
	return perms.Can(current.Role, anyPermission)
}

func canAccessStudentOwner(current model.AuthUser, ownerID *int, perms *helper.PermissionSet, anyPermission string) bool {
	if ownerID == nil {
		return perms.Can(current.Role, anyPermission)
	}
	return CanAccessStudent(current, *ownerID, perms, anyPermission)
}
