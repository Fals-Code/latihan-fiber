package helper

import "testing"

func TestPermissionSetAllowsKnownPermission(t *testing.T) {
	perms := NewPermissionSet(map[string][]string{
		"admin": {"user:list", "user:read:any"},
		"staff": {"user:list"},
		"user":  {},
	})

	if !perms.Can("admin", "user:list") {
		t.Fatal("expected admin to have user:list")
	}
	if !perms.Can("staff", "user:list") {
		t.Fatal("expected staff to have user:list")
	}
}

func TestPermissionSetDeniesUnknownAndMissingPermissions(t *testing.T) {
	perms := NewPermissionSet(map[string][]string{"admin": {"user:list"}, "user": {}})

	for _, tc := range []struct {
		role       string
		permission string
	}{
		{"admin", "user:delete"},
		{"unknown", "user:list"},
		{"admin", "unknown"},
	} {
		if perms.Can(tc.role, tc.permission) {
			t.Fatalf("expected denial for role=%q permission=%q", tc.role, tc.permission)
		}
	}
}

func TestPermissionSetNilIsFailClosed(t *testing.T) {
	var perms *PermissionSet
	if perms.Can("admin", "user:list") {
		t.Fatal("nil PermissionSet must deny")
	}
	if got := perms.PermissionsOf("admin"); len(got) != 0 {
		t.Fatalf("expected empty permissions, got %v", got)
	}
	if got := perms.KnownRoles(); len(got) != 0 {
		t.Fatalf("expected empty roles, got %v", got)
	}
	if perms.IsKnownRole("admin") {
		t.Fatal("nil PermissionSet must not know roles")
	}
}

func TestPermissionSetPreservesKnownRoleWithoutPermissions(t *testing.T) {
	perms := NewPermissionSet(map[string][]string{"admin": {"user:list"}, "user": {}})

	if !perms.IsKnownRole("user") {
		t.Fatal("user must remain a known role")
	}
	if perms.Can("user", "user:list") {
		t.Fatal("user must have no permissions")
	}
	if got := perms.PermissionsOf("user"); got == nil || len(got) != 0 {
		t.Fatalf("expected non-nil empty permissions, got %v", got)
	}
}

func TestPermissionSetReturnsSortedValues(t *testing.T) {
	perms := NewPermissionSet(map[string][]string{
		"staff": {"user:read:any", "user:list"},
		"admin": {"role:assign"},
		"user":  {},
	})

	if got, want := perms.PermissionsOf("staff"), []string{"user:list", "user:read:any"}; !equalStrings(got, want) {
		t.Fatalf("permissions = %v, want %v", got, want)
	}
	if got, want := perms.KnownRoles(), []string{"admin", "staff", "user"}; !equalStrings(got, want) {
		t.Fatalf("roles = %v, want %v", got, want)
	}
}

func TestPermissionSetIsKnownRole(t *testing.T) {
	perms := NewPermissionSet(map[string][]string{"user": {}})
	if !perms.IsKnownRole("user") {
		t.Fatal("expected user to be known")
	}
	if perms.IsKnownRole("admin") {
		t.Fatal("did not expect admin to be known")
	}
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
