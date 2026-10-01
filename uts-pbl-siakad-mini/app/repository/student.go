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

func (r *StudentRepository) GetDetail(ctx context.Context, id int64) (model.StudentDetail, error) {
	var detail model.StudentDetail
	err := r.pool.QueryRow(ctx, "SELECT id, user_id, nim, nama, prodi, angkatan, ipk_terakhir FROM students WHERE id = $1 AND deleted_at IS NULL", id).Scan(&detail.ID, &detail.UserID, &detail.NIM, &detail.Nama, &detail.Prodi, &detail.Angkatan, &detail.IPKTerakhir)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.StudentDetail{}, ErrNotFound
	}
	if err != nil {
		return model.StudentDetail{}, err
	}
	rows, err := r.pool.Query(ctx, "SELECT e.id, c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, e.tahun_akademik FROM enrollments e JOIN courses c ON c.id = e.course_id WHERE e.student_id = $1 ORDER BY e.id", id)
	if err != nil {
		return model.StudentDetail{}, err
	}
	defer rows.Close()
	detail.Courses = make([]model.StudentCourse, 0)
	for rows.Next() {
		var course model.StudentCourse
		if err := rows.Scan(&course.EnrollmentID, &course.CourseID, &course.KodeMK, &course.NamaMK, &course.SKS, &course.Semester, &course.TahunAkademik); err != nil {
			return model.StudentDetail{}, err
		}
		detail.Courses = append(detail.Courses, course)
		detail.TotalSKS += course.SKS
	}
	if err := rows.Err(); err != nil {
		return model.StudentDetail{}, err
	}
	return detail, nil
}

func (r *StudentRepository) StudentIDByUserID(ctx context.Context, userID int64) (int64, error) {
	var studentID int64
	err := r.pool.QueryRow(ctx, "SELECT id FROM students WHERE user_id = $1 AND deleted_at IS NULL", userID).Scan(&studentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNotFound
	}
	return studentID, err
}

func (r *StudentRepository) Update(ctx context.Context, id int64, update model.StudentUpdate) (model.Student, error) {
	sets := make([]string, 0, 4)
	args := make([]any, 0, 5)
	add := func(column string, value any) {
		args = append(args, value)
		sets = append(sets, fmt.Sprintf("%s = $%d", column, len(args)))
	}
	if update.Nama != nil {
		add("nama", *update.Nama)
	}
	if update.Prodi != nil {
		add("prodi", *update.Prodi)
	}
	if update.Angkatan != nil {
		add("angkatan", *update.Angkatan)
	}
	if update.IPKTerakhir != nil {
		add("ipk_terakhir", *update.IPKTerakhir)
	}
	if len(sets) == 0 {
		return model.Student{}, errors.New("empty student update")
	}
	args = append(args, id)
	query := "UPDATE students SET " + strings.Join(sets, ", ") + fmt.Sprintf(", updated_at = CURRENT_TIMESTAMP WHERE id = $%d AND deleted_at IS NULL RETURNING id, nim, nama, prodi, angkatan, ipk_terakhir", len(args))
	var student model.Student
	err := r.pool.QueryRow(ctx, query, args...).Scan(&student.ID, &student.NIM, &student.Nama, &student.Prodi, &student.Angkatan, &student.IPKTerakhir)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Student{}, ErrNotFound
	}
	return student, err
}

func (r *StudentRepository) SoftDelete(ctx context.Context, id int64) error {
	var deletedID int64
	err := r.pool.QueryRow(ctx, "UPDATE students SET deleted_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = $1 AND deleted_at IS NULL RETURNING id", id).Scan(&deletedID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
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
