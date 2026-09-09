package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"tugas1-go/pertemuan-5-authentication-security/app/model"
)

type UserRepository interface {
	Create(ctx context.Context, user model.User) (model.User, error)
	FindByUsername(ctx context.Context, username string) (model.User, error)
	FindByID(ctx context.Context, id int) (model.User, error)
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

const userColumns = `id, username, email, password_hash, role, is_active, created_at`

func (r *userRepository) Create(ctx context.Context, user model.User) (model.User, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO users (username, email, password_hash, role, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, username, email, password_hash, role, is_active, created_at
	`, user.Username, user.Email, user.PasswordHash, user.Role, user.IsActive).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.IsActive, &user.CreatedAt,
	)
	if err != nil {
		return model.User{}, mapUserDatabaseError("create user", err)
	}
	return user, nil
}

func (r *userRepository) FindByUsername(ctx context.Context, username string) (model.User, error) {
	var user model.User
	err := r.db.QueryRow(ctx, `
		SELECT id, username, email, password_hash, role, is_active, created_at
		FROM users
		WHERE LOWER(username) = LOWER($1)
	`, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.IsActive, &user.CreatedAt,
	)
	if err != nil {
		return model.User{}, mapUserDatabaseError("find user by username", err)
	}
	return user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id int) (model.User, error) {
	var user model.User
	err := r.db.QueryRow(ctx, fmt.Sprintf("SELECT %s FROM users WHERE id = $1", userColumns), id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.Role, &user.IsActive, &user.CreatedAt,
	)
	if err != nil {
		return model.User{}, mapUserDatabaseError("find user by id", err)
	}
	return user, nil
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
