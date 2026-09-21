package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"tugas1-go/pertemuan-7-advanced-api-design/app/model"
)

type UserRepository interface {
	Create(ctx context.Context, user model.User) (model.User, error)
	FindByUsername(ctx context.Context, username string) (model.User, error)
	FindByID(ctx context.Context, id int) (model.User, error)
	FindAll(ctx context.Context) ([]model.User, error)
	Update(ctx context.Context, user model.User) (model.User, error)
	UpdateRole(ctx context.Context, id int, role string) (model.User, error)
	Delete(ctx context.Context, id int) error
}

type userRepository struct{ db *pgxpool.Pool }

func NewUserRepository(db *pgxpool.Pool) UserRepository { return &userRepository{db: db} }

const userColumns = `id, username, email, password_hash, role, is_active, created_at`

func (r *userRepository) Create(ctx context.Context, user model.User) (model.User, error) {
	err := r.db.QueryRow(ctx, `INSERT INTO users (username, email, password_hash, role, is_active) VALUES ($1, $2, $3, $4, $5) RETURNING id, username, email, password_hash, role, is_active, created_at`, user.Username, user.Email, user.PasswordHash, user.Role, user.IsActive).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.IsActive, &user.CreatedAt)
	if err != nil {
		return model.User{}, mapUserDatabaseError("create user", err)
	}
	return user, nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (model.User, error) {
	var user model.User
	err := r.db.QueryRow(ctx, `SELECT id, username, email, password_hash, role, is_active, created_at FROM users WHERE LOWER(username) = LOWER($1)`, username).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.IsActive, &user.CreatedAt)
	if err != nil {
		return model.User{}, mapUserDatabaseError("find user by username", err)
	}
	return user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id int) (model.User, error) {
	var user model.User
	err := r.db.QueryRow(ctx, fmt.Sprintf("SELECT %s FROM users WHERE id = $1", userColumns), id).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.IsActive, &user.CreatedAt)
	if err != nil {
		return model.User{}, mapUserDatabaseError("find user by id", err)
	}
	return user, nil
}

func (r *userRepository) FindAll(ctx context.Context) ([]model.User, error) {
	rows, err := r.db.Query(ctx, fmt.Sprintf("SELECT %s FROM users ORDER BY id", userColumns))
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()
	users := make([]model.User, 0)
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.IsActive, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}
	return users, nil
}

func (r *userRepository) Update(ctx context.Context, user model.User) (model.User, error) {
	err := r.db.QueryRow(ctx, `UPDATE users SET username = $1, email = $2, is_active = $3 WHERE id = $4 RETURNING id, username, email, password_hash, role, is_active, created_at`, user.Username, user.Email, user.IsActive, user.ID).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.IsActive, &user.CreatedAt)
	if err != nil {
		return model.User{}, mapUserDatabaseError("update user", err)
	}
	return user, nil
}

func (r *userRepository) UpdateRole(ctx context.Context, id int, role string) (model.User, error) {
	var user model.User
	err := r.db.QueryRow(ctx, `UPDATE users SET role = $1 WHERE id = $2 RETURNING id, username, email, password_hash, role, is_active, created_at`, role, id).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.IsActive, &user.CreatedAt)
	if err != nil {
		return model.User{}, mapUserDatabaseError("update user role", err)
	}
	return user, nil
}

func (r *userRepository) Delete(ctx context.Context, id int) error {
	result, err := r.db.Exec(ctx, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func mapUserDatabaseError(operation string, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrDuplicate
	}
	return fmt.Errorf("%s: %w", operation, err)
}
