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

var ErrStudentNotFound = errors.New("student tidak ditemukan")

type AchievementRepository interface {
	FindAll(ctx context.Context, q model.ListQueryAchievement) ([]model.Achievement, int, error)
	FindByID(ctx context.Context, id int) (model.Achievement, error)
	Create(ctx context.Context, achievement model.Achievement) (model.Achievement, error)
	Update(ctx context.Context, achievement model.Achievement) (model.Achievement, error)
	Delete(ctx context.Context, id int) error
}

type achievementRepository struct {
	db *pgxpool.Pool
}

func NewAchievementRepository(db *pgxpool.Pool) AchievementRepository {
	return &achievementRepository{db: db}
}

var allowedAchievementSortColumns = map[string]string{
	"id":         "id",
	"name":       "name",
	"student_id": "student_id",
	"rank":       "rank",
	"created_at": "created_at",
}

func buildAchievementFilter(q model.ListQueryAchievement) (string, []any) {
	if q.Search == "" {
		return "", nil
	}

	return " WHERE name ILIKE $1", []any{"%" + q.Search + "%"}
}

func (r *achievementRepository) FindAll(ctx context.Context, q model.ListQueryAchievement) ([]model.Achievement, int, error) {
	whereClause, args := buildAchievementFilter(q)

	var total int
	if err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM achievements"+whereClause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	sortColumn := allowedAchievementSortColumns[q.Sort]
	if sortColumn == "" {
		sortColumn = "id"
	}
	order := "ASC"
	if q.Order == "desc" {
		order = "DESC"
	}

	limitPosition := len(args) + 1
	offsetPosition := len(args) + 2
	query := fmt.Sprintf(`
		SELECT id, name, student_id, rank, created_at
		FROM achievements
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortColumn, order, limitPosition, offsetPosition)
	queryArgs := append(append([]any{}, args...), q.Limit, q.Offset())

	rows, err := r.db.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	achievements := make([]model.Achievement, 0)
	for rows.Next() {
		var achievement model.Achievement
		if err := rows.Scan(&achievement.ID, &achievement.Name, &achievement.StudentID, &achievement.Rank, &achievement.CreatedAt); err != nil {
			return nil, 0, err
		}
		achievements = append(achievements, achievement)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return achievements, total, nil
}

func (r *achievementRepository) FindByID(ctx context.Context, id int) (model.Achievement, error) {
	var achievement model.Achievement
	err := r.db.QueryRow(ctx, `
		SELECT id, name, student_id, rank, created_at
		FROM achievements WHERE id = $1
	`, id).Scan(&achievement.ID, &achievement.Name, &achievement.StudentID, &achievement.Rank, &achievement.CreatedAt)
	if err != nil {
		return model.Achievement{}, mapAchievementDatabaseError(err)
	}
	return achievement, nil
}

func (r *achievementRepository) Create(ctx context.Context, achievement model.Achievement) (model.Achievement, error) {
	err := r.db.QueryRow(ctx, `
		INSERT INTO achievements (name, student_id, rank)
		VALUES ($1, $2, $3)
		RETURNING id, name, student_id, rank, created_at
	`, achievement.Name, achievement.StudentID, achievement.Rank).Scan(
		&achievement.ID, &achievement.Name, &achievement.StudentID, &achievement.Rank, &achievement.CreatedAt,
	)
	if err != nil {
		return model.Achievement{}, mapAchievementDatabaseError(err)
	}
	return achievement, nil
}

func (r *achievementRepository) Update(ctx context.Context, achievement model.Achievement) (model.Achievement, error) {
	err := r.db.QueryRow(ctx, `
		UPDATE achievements
		SET name = $1, student_id = $2, rank = $3
		WHERE id = $4
		RETURNING id, name, student_id, rank, created_at
	`, achievement.Name, achievement.StudentID, achievement.Rank, achievement.ID).Scan(
		&achievement.ID, &achievement.Name, &achievement.StudentID, &achievement.Rank, &achievement.CreatedAt,
	)
	if err != nil {
		return model.Achievement{}, mapAchievementDatabaseError(err)
	}
	return achievement, nil
}

func (r *achievementRepository) Delete(ctx context.Context, id int) error {
	result, err := r.db.Exec(ctx, "DELETE FROM achievements WHERE id = $1", id)
	if err != nil {
		return mapAchievementDatabaseError(err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func mapAchievementDatabaseError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return ErrStudentNotFound
	}
	return err
}
