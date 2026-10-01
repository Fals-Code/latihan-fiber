package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"tugas1-go/uts-pbl-siakad-mini/app/model"
)

type CourseRepository struct{ pool *pgxpool.Pool }

func NewCourseRepository(pool *pgxpool.Pool) *CourseRepository {
	return &CourseRepository{pool: pool}
}

func (r *CourseRepository) List(ctx context.Context, filters model.CourseFilters) ([]model.Course, error) {
	conditions := make([]string, 0, 3)
	args := make([]any, 0, 3)
	addArg := func(value any) string {
		args = append(args, value)
		return fmt.Sprintf("$%d", len(args))
	}
	if filters.Semester != nil {
		conditions = append(conditions, "c.semester = "+addArg(*filters.Semester))
	}
	if filters.Search != "" {
		placeholder := addArg("%" + filters.Search + "%")
		conditions = append(conditions, "(c.kode_mk ILIKE "+placeholder+" OR c.nama_mk ILIKE "+placeholder+")")
	}
	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}
	query := "SELECT c.id, c.kode_mk, c.nama_mk, c.sks, c.semester, c.kuota, COUNT(e.id)::BIGINT AS terisi, (c.kuota - COUNT(e.id))::BIGINT AS sisa_kuota FROM courses c LEFT JOIN enrollments e ON e.course_id = c.id" + where + " GROUP BY c.id"
	if filters.Available {
		query += " HAVING COUNT(e.id) < c.kuota"
	}
	query += " ORDER BY c.id ASC"
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	courses := make([]model.Course, 0)
	for rows.Next() {
		var course model.Course
		if err := rows.Scan(&course.ID, &course.KodeMK, &course.NamaMK, &course.SKS, &course.Semester, &course.Kuota, &course.Terisi, &course.SisaKuota); err != nil {
			return nil, err
		}
		courses = append(courses, course)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return courses, nil
}
