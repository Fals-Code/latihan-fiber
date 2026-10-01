package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"tugas1-go/uts-pbl-siakad-mini/app/model"
)

var ErrNotFound = errors.New("record not found")

type AuthRepository struct {
	pool *pgxpool.Pool
}

func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{pool: pool}
}

func (r *AuthRepository) FindUserByEmail(ctx context.Context, email string) (model.User, error) {
	var user model.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password, role FROM users WHERE email = $1`, email,
	).Scan(&user.ID, &user.Email, &user.Password, &user.Role)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.User{}, ErrNotFound
	}
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (r *AuthRepository) FindActiveAccount(ctx context.Context, userID int64) (model.CurrentUser, error) {
	var user model.CurrentUser
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, role FROM users WHERE id = $1`, userID,
	).Scan(&user.ID, &user.Email, &user.Role)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.CurrentUser{}, ErrNotFound
	}
	if err != nil {
		return model.CurrentUser{}, err
	}
	if user.Role == "mahasiswa" {
		var student model.StudentProfile
		err := r.pool.QueryRow(ctx,
			`SELECT nim, nama, prodi, angkatan FROM students WHERE user_id = $1 AND deleted_at IS NULL`, userID,
		).Scan(&student.NIM, &student.Nama, &student.Prodi, &student.Angkatan)
		if errors.Is(err, pgx.ErrNoRows) {
			return model.CurrentUser{}, ErrNotFound
		}
		if err != nil {
			return model.CurrentUser{}, err
		}
		user.Student = &student
	}
	if user.Role != "admin" && user.Role != "mahasiswa" {
		return model.CurrentUser{}, ErrNotFound
	}
	return user, nil
}
