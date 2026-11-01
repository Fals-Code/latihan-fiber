package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RoleRepository interface {
	LoadPermissions(ctx context.Context) (map[string][]string, error)
}

type roleRepository struct {
	db *pgxpool.Pool
}

func NewRoleRepository(db *pgxpool.Pool) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) LoadPermissions(ctx context.Context) (map[string][]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT r.name, COALESCE(rp.permission_name, '')
		FROM roles r
		LEFT JOIN role_permissions rp ON rp.role_name = r.name
		ORDER BY r.name, rp.permission_name
	`)
	if err != nil {
		return nil, fmt.Errorf("load role permissions: %w", err)
	}
	defer rows.Close()

	permissions := make(map[string][]string)
	for rows.Next() {
		var role, permission string
		if err := rows.Scan(&role, &permission); err != nil {
			return nil, fmt.Errorf("scan role permissions: %w", err)
		}
		if _, ok := permissions[role]; !ok {
			permissions[role] = []string{}
		}
		if permission != "" {
			permissions[role] = append(permissions[role], permission)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate role permissions: %w", err)
	}

	return permissions, nil
}
