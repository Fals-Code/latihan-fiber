package helper

import "sort"

type PermissionSet struct {
	byRole map[string]map[string]struct{}
}

func NewPermissionSet(raw map[string][]string) *PermissionSet {
	perms := &PermissionSet{byRole: make(map[string]map[string]struct{}, len(raw))}
	for role, permissions := range raw {
		set := make(map[string]struct{}, len(permissions))
		for _, permission := range permissions {
			if permission != "" {
				set[permission] = struct{}{}
			}
		}
		perms.byRole[role] = set
	}
	return perms
}

func (p *PermissionSet) Can(role, permission string) bool {
	if p == nil || permission == "" {
		return false
	}
	permissions, ok := p.byRole[role]
	if !ok {
		return false
	}
	_, ok = permissions[permission]
	return ok
}

func (p *PermissionSet) PermissionsOf(role string) []string {
	if p == nil {
		return []string{}
	}
	permissions, ok := p.byRole[role]
	if !ok {
		return []string{}
	}
	result := make([]string, 0, len(permissions))
	for permission := range permissions {
		result = append(result, permission)
	}
	sort.Strings(result)
	return result
}

func (p *PermissionSet) KnownRoles() []string {
	if p == nil {
		return []string{}
	}
	roles := make([]string, 0, len(p.byRole))
	for role := range p.byRole {
		roles = append(roles, role)
	}
	sort.Strings(roles)
	return roles
}

func (p *PermissionSet) IsKnownRole(role string) bool {
	if p == nil {
		return false
	}
	_, ok := p.byRole[role]
	return ok
}
