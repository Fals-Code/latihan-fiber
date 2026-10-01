package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"tugas1-go/uts-pbl-siakad-mini/app/model"
)

var ErrDuplicateNIM = errors.New("duplicate student NIM")
var ErrDuplicateEmail = errors.New("duplicate student email")

type StudentRepository struct {
	pool *pgxpool.Pool
}

func NewStudentRepository(pool *pgxpool.Pool) *StudentRepository {
	return &StudentRepository{pool: pool}
}

func (r *StudentRepository) List(ctx context.Context, filters model.StudentFilters) ([]model.Student, int64, error) {
	conditions := []string{"s.deleted_at IS NULL"}
	args := make([]any, 0, 4)
	addArg := func(value any) string {
		args = append(args, value)
		return fmt.Sprintf("$%d", len(args))
	}
	if filters.Prodi != "" {
		conditions = append(conditions, "s.prodi = "+addArg(filters.Prodi))
	}
	if filters.Angkatan != nil {
		conditions = append(conditions, "s.angkatan = "+addArg(*filters.Angkatan))
	}
	if filters.Search != "" {
		placeholder := addArg("%" + filters.Search + "%")
		conditions = append(conditions, "(s.nim ILIKE "+placeholder+" OR s.nama ILIKE "+placeholder+")")
	}
	where := strings.Join(conditions, " AND ")
	var total int64
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM students s WHERE "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	orderBy := "s.id ASC"
	switch filters.Sort {
	case "nama":
		orderBy = "s.nama ASC, s.id ASC"
	case "-ipk_terakhir":
		orderBy = "s.ipk_terakhir DESC, s.id ASC"
	}
	listArgs := append(append([]any(nil), args...), int64(filters.PerPage), (int64(filters.Page)-1)*int64(filters.PerPage))
	query := "SELECT s.id, s.nim, s.nama, s.prodi, s.angkatan, s.ipk_terakhir FROM students s WHERE " + where + " ORDER BY " + orderBy + fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(listArgs)-1, len(listArgs))
	rows, err := r.pool.Query(ctx, query, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	students := make([]model.Student, 0)
	for rows.Next() {
		var student model.Student
		if err := rows.Scan(&student.ID, &student.NIM, &student.Nama, &student.Prodi, &student.Angkatan, &student.IPKTerakhir); err != nil {
			return nil, 0, err
		}
		students = append(students, student)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return students, total, nil
}

type studentTransaction interface {
	QueryRow(context.Context, string, ...any) pgx.Row
	Commit(context.Context) error
	Rollback(context.Context) error
}

type studentTxStarter func(context.Context, pgx.TxOptions) (studentTransaction, error)

func (r *StudentRepository) Create(ctx context.Context, student model.NewStudent, hashPassword func() (string, error)) (model.Student, error) {
	starter := studentTxStarter(func(ctx context.Context, options pgx.TxOptions) (studentTransaction, error) {
		return r.pool.BeginTx(ctx, options)
	})
	return createStudentTx(ctx, starter, student, hashPassword)
}

func createStudentTx(ctx context.Context, starter studentTxStarter, student model.NewStudent, hashPassword func() (string, error)) (model.Student, error) {
	tx, err := starter(ctx, pgx.TxOptions{})
	if err != nil {
		return model.Student{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var exists bool
	if err := tx.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM students WHERE nim = $1)", student.NIM).Scan(&exists); err != nil {
		return model.Student{}, err
	}
	if exists {
		return model.Student{}, ErrDuplicateNIM
	}
	if err := tx.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM users WHERE email = $1)", student.Email).Scan(&exists); err != nil {
		return model.Student{}, err
	}
	if exists {
		return model.Student{}, ErrDuplicateEmail
	}
	student.Password, err = hashPassword()
	if err != nil {
		return model.Student{}, err
	}

	var userID int64
	err = tx.QueryRow(ctx, "INSERT INTO users (email, password, role) VALUES ($1, $2, 'mahasiswa') RETURNING id", student.Email, student.Password).Scan(&userID)
	if err != nil {
		return model.Student{}, duplicateError(err)
	}
	var created model.Student
	err = tx.QueryRow(ctx, "INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, nim, nama, prodi, angkatan, ipk_terakhir", userID, student.NIM, student.Nama, student.Prodi, student.Angkatan, student.IPKTerakhir).Scan(&created.ID, &created.NIM, &created.Nama, &created.Prodi, &created.Angkatan, &created.IPKTerakhir)
	if err != nil {
		return model.Student{}, duplicateError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return model.Student{}, duplicateError(err)
	}
	return created, nil
}

func duplicateError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return err
	}
	switch pgErr.ConstraintName {
	case "users_email_key":
		return ErrDuplicateEmail
	case "students_nim_key":
		return ErrDuplicateNIM
	default:
		return err
	}
}
