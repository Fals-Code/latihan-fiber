package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"tugas1-go/uts-pbl-siakad-mini/app/model"
)

var (
	ErrDuplicateEnrollment = errors.New("course already enrolled for academic year")
	ErrEnrollmentQuotaFull = errors.New("course quota is full")
	ErrEnrollmentSKSLimit  = errors.New("enrollment exceeds SKS limit")
	ErrInvalidAcademicYear = errors.New("invalid academic year")
	ErrForbidden           = errors.New("record belongs to another student")
)

var academicYearPattern = regexp.MustCompile(`^[0-9]{4}/[0-9]{4}-(Ganjil|Genap)$`)

type EnrollmentRepository struct{ pool *pgxpool.Pool }

func NewEnrollmentRepository(pool *pgxpool.Pool) *EnrollmentRepository {
	return &EnrollmentRepository{pool: pool}
}

func (r *EnrollmentRepository) Create(ctx context.Context, userID int64, enrollment model.NewEnrollment) (model.Enrollment, error) {
	if !academicYearPattern.MatchString(enrollment.TahunAkademik) || enrollment.CourseID < 1 {
		return model.Enrollment{}, ErrInvalidAcademicYear
	}
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return model.Enrollment{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var studentID int64
	var ipk float64
	err = tx.QueryRow(ctx, "SELECT id, ipk_terakhir FROM students WHERE user_id = $1 AND deleted_at IS NULL FOR UPDATE", userID).Scan(&studentID, &ipk)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Enrollment{}, ErrForbidden
	}
	if err != nil {
		return model.Enrollment{}, err
	}
	var course model.Course
	err = tx.QueryRow(ctx, "SELECT id, sks, kuota FROM courses WHERE id = $1 FOR UPDATE", enrollment.CourseID).Scan(&course.ID, &course.SKS, &course.Kuota)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Enrollment{}, ErrNotFound
	}
	if err != nil {
		return model.Enrollment{}, err
	}
	var exists bool
	err = tx.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM enrollments WHERE student_id = $1 AND course_id = $2 AND tahun_akademik = $3)", studentID, course.ID, enrollment.TahunAkademik).Scan(&exists)
	if err != nil {
		return model.Enrollment{}, err
	}
	if exists {
		return model.Enrollment{}, ErrDuplicateEnrollment
	}
	var enrolled int64
	err = tx.QueryRow(ctx, "SELECT COUNT(*) FROM enrollments WHERE course_id = $1", course.ID).Scan(&enrolled)
	if err != nil {
		return model.Enrollment{}, err
	}
	if enrolled >= int64(course.Kuota) {
		return model.Enrollment{}, ErrEnrollmentQuotaFull
	}
	var totalSKS int
	err = tx.QueryRow(ctx, "SELECT COALESCE(SUM(c.sks), 0) FROM enrollments e JOIN courses c ON c.id = e.course_id WHERE e.student_id = $1 AND e.tahun_akademik = $2", studentID, enrollment.TahunAkademik).Scan(&totalSKS)
	if err != nil {
		return model.Enrollment{}, err
	}
	limit := 18
	if ipk >= 3 {
		limit = 24
	} else if ipk >= 2.5 {
		limit = 21
	}
	if totalSKS+course.SKS > limit {
		return model.Enrollment{}, fmt.Errorf("%w: sisa SKS yang dapat diambil %d", ErrEnrollmentSKSLimit, limit-totalSKS)
	}
	created := model.Enrollment{StudentID: studentID, CourseID: course.ID, TahunAkademik: enrollment.TahunAkademik}
	err = tx.QueryRow(ctx, "INSERT INTO enrollments (student_id, course_id, tahun_akademik) VALUES ($1, $2, $3) RETURNING id, created_at", studentID, course.ID, enrollment.TahunAkademik).Scan(&created.ID, &created.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.Enrollment{}, ErrDuplicateEnrollment
		}
		return model.Enrollment{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return model.Enrollment{}, err
	}
	return created, nil
}

func (r *EnrollmentRepository) Delete(ctx context.Context, userID, enrollmentID int64) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var studentID int64
	err = tx.QueryRow(ctx, "SELECT id FROM students WHERE user_id = $1 AND deleted_at IS NULL FOR UPDATE", userID).Scan(&studentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrForbidden
	}
	if err != nil {
		return err
	}
	var owner int64
	err = tx.QueryRow(ctx, "SELECT student_id FROM enrollments WHERE id = $1 FOR UPDATE", enrollmentID).Scan(&owner)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if owner != studentID {
		return ErrForbidden
	}
	if _, err := tx.Exec(ctx, "DELETE FROM enrollments WHERE id = $1 AND student_id = $2", enrollmentID, studentID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
