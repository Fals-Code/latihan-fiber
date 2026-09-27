package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"tugas1-go/pertemuan-7-advanced-api-design/app/model"
)

var (
	ErrNotFound  = errors.New("data tidak ditemukan")
	ErrDuplicate = errors.New("data duplikat")
)

type StudentRepository interface {
	FindAll(ctx context.Context, q model.ListQuery) ([]model.Student, int, error)
	FindFirstPage(ctx context.Context, q model.ListQuery) ([]model.Student, error)
	FindAfterCursor(ctx context.Context, q model.ListQuery, createdAt time.Time, id int) ([]model.Student, error)
	FindByID(ctx context.Context, id int) (model.Student, error)
	FindOwnerIDByID(ctx context.Context, id int) (*int, error)
	Create(ctx context.Context, student model.Student) (model.Student, error)
	Update(ctx context.Context, student model.Student) (model.Student, error)
	Delete(ctx context.Context, id int) error
}

type studentRepository struct {
	db *pgxpool.Pool
}

func NewStudentRepository(db *pgxpool.Pool) StudentRepository {
	return &studentRepository{
		db: db,
	}
}

var allowedSortColumns = map[string]string{
	"id":         "id",
	"nim":        "nim",
	"name":       "name",
	"grade":      "grade",
	"is_active":  "is_active",
	"created_at": "created_at",
}

func buildStudentFilter(q model.ListQuery) (string, []any) {
	conditions := []string{}
	args := []any{}

	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		conditions = append(
			conditions,
			fmt.Sprintf("name ILIKE $%d", len(args)),
		)
	}

	if q.IsActive != nil {
		args = append(args, *q.IsActive)
		conditions = append(
			conditions,
			fmt.Sprintf("is_active = $%d", len(args)),
		)
	}

	if len(conditions) == 0 {
		return "", args
	}

	return " WHERE " + strings.Join(conditions, " AND "), args
}

func (r *studentRepository) FindAll(
	ctx context.Context,
	q model.ListQuery,
) ([]model.Student, int, error) {
	whereClause, args := buildStudentFilter(q)

	countQuery := `
		SELECT COUNT(*)
		FROM students
	` + whereClause

	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limitPosition := len(args) + 1
	offsetPosition := len(args) + 2

	query := fmt.Sprintf(`
		SELECT id, nim, name, grade, is_active, owner_id, created_at
		FROM students
		%s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, limitPosition, offsetPosition)

	queryArgs := append(
		append([]any{}, args...),
		q.Limit+1,
		q.Offset(),
	)

	rows, err := r.db.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	students := make([]model.Student, 0)

	for rows.Next() {
		var student model.Student

		if err := rows.Scan(
			&student.ID,
			&student.NIM,
			&student.Name,
			&student.Grade,
			&student.IsActive,
			&student.OwnerID,
			&student.CreatedAt,
		); err != nil {
			return nil, 0, err
		}

		students = append(students, student)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return students, total, nil
}

func buildStudentFirstPageQuery(q model.ListQuery) (string, []any) {
	whereClause, filterArgs := buildStudentFilter(q)
	limitPosition := len(filterArgs) + 1
	query := fmt.Sprintf("SELECT id, nim, name, grade, is_active, owner_id, created_at FROM students%s ORDER BY created_at DESC, id DESC LIMIT $%d", whereClause, limitPosition)
	args := append([]any{}, filterArgs...)
	args = append(args, q.Limit+1)
	return query, args
}

func (r *studentRepository) FindFirstPage(ctx context.Context, q model.ListQuery) ([]model.Student, error) {
	query, args := buildStudentFirstPageQuery(q)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	students := make([]model.Student, 0)
	for rows.Next() {
		var student model.Student
		if err := rows.Scan(&student.ID, &student.NIM, &student.Name, &student.Grade, &student.IsActive, &student.OwnerID, &student.CreatedAt); err != nil {
			return nil, err
		}
		students = append(students, student)
	}
	return students, rows.Err()
}

func buildStudentCursorQuery(q model.ListQuery, createdAt time.Time, id int) (string, []any) {
	filter, filterArgs := buildStudentFilter(q)
	cursorPosition := len(filterArgs) + 1
	limitPosition := cursorPosition + 2
	whereClause := fmt.Sprintf(" WHERE (created_at, id) < ($%d, $%d)", cursorPosition, cursorPosition+1)
	if filter != "" {
		whereClause = filter + " AND" + whereClause[7:]
	}
	args := append([]any{}, filterArgs...)
	args = append(args, createdAt, id, q.Limit+1)
	query := fmt.Sprintf("SELECT id, nim, name, grade, is_active, owner_id, created_at FROM students%s ORDER BY created_at DESC, id DESC LIMIT $%d", whereClause, limitPosition)
	return query, args
}

func (r *studentRepository) FindAfterCursor(ctx context.Context, q model.ListQuery, createdAt time.Time, id int) ([]model.Student, error) {
	query, args := buildStudentCursorQuery(q, createdAt, id)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.Student, 0)
	for rows.Next() {
		var item model.Student
		if err := rows.Scan(&item.ID, &item.NIM, &item.Name, &item.Grade, &item.IsActive, &item.OwnerID, &item.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *studentRepository) FindByID(
	ctx context.Context,
	id int,
) (model.Student, error) {
	var student model.Student

	err := r.db.QueryRow(
		ctx,
		`
			SELECT id, nim, name, grade, is_active, owner_id, created_at
			FROM students
			WHERE id = $1
		`,
		id,
	).Scan(
		&student.ID,
		&student.NIM,
		&student.Name,
		&student.Grade,
		&student.IsActive,
		&student.OwnerID,
		&student.CreatedAt,
	)

	if err != nil {
		return model.Student{}, mapDatabaseError(err)
	}

	return student, nil
}

func (r *studentRepository) FindOwnerIDByID(ctx context.Context, id int) (*int, error) {
	var ownerID *int
	if err := r.db.QueryRow(ctx, `SELECT owner_id FROM students WHERE id = $1`, id).Scan(&ownerID); err != nil {
		return nil, mapDatabaseError(err)
	}
	return ownerID, nil
}

func (r *studentRepository) Create(
	ctx context.Context,
	student model.Student,
) (model.Student, error) {
	err := r.db.QueryRow(
		ctx,
		`
			INSERT INTO students (nim, name, grade, is_active, owner_id)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, nim, name, grade, is_active, owner_id, created_at
		`,
		student.NIM,
		student.Name,
		student.Grade,
		student.IsActive,
		student.OwnerID,
	).Scan(
		&student.ID,
		&student.NIM,
		&student.Name,
		&student.Grade,
		&student.IsActive,
		&student.OwnerID,
		&student.CreatedAt,
	)

	if err != nil {
		return model.Student{}, mapDatabaseError(err)
	}

	return student, nil
}

func (r *studentRepository) Update(
	ctx context.Context,
	student model.Student,
) (model.Student, error) {
	err := r.db.QueryRow(
		ctx,
		`
			UPDATE students
			SET nim = $1,
			    name = $2,
			    grade = $3,
			    is_active = $4
			WHERE id = $5
			RETURNING id, nim, name, grade, is_active, owner_id, created_at
		`,
		student.NIM,
		student.Name,
		student.Grade,
		student.IsActive,
		student.ID,
	).Scan(
		&student.ID,
		&student.NIM,
		&student.Name,
		&student.Grade,
		&student.IsActive,
		&student.OwnerID,
		&student.CreatedAt,
	)

	if err != nil {
		return model.Student{}, mapDatabaseError(err)
	}

	return student, nil
}

func (r *studentRepository) Delete(
	ctx context.Context,
	id int,
) error {
	result, err := r.db.Exec(
		ctx,
		`
			DELETE FROM students
			WHERE id = $1
		`,
		id,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func mapDatabaseError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return ErrDuplicate
		}
	}

	return err
}
