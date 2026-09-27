package repository

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"tugas1-go/pertemuan-7-advanced-api-design/app/model"
)

const runDBIntegrationTests = "RUN_DB_INTEGRATION_TESTS"

func TestCursorQueriesAgainstPostgres(t *testing.T) {
	if os.Getenv(runDBIntegrationTests) != "1" {
		t.Skip(runDBIntegrationTests + "=1 is required; skipping before loading .env or opening a database connection")
	}
	if err := godotenv.Load("../../.env"); err != nil {
		t.Fatalf("load database configuration: %v", err)
	}
	user := getenvOrDefault("DB_USER", "postgres")
	password := os.Getenv("DB_PASSWORD")
	host := getenvOrDefault("DB_HOST", "localhost")
	port := getenvOrDefault("DB_PORT", "5432")
	name := os.Getenv("DB_NAME")
	requireTestDatabase(t, name)
	sslmode := getenvOrDefault("DB_SSLMODE", "disable")
	dsn := buildPostgresDSN(user, password, host, port, name, sslmode)
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("database pool: %v", err)
	}
	defer pool.Close()
	ctx := context.Background()
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("database ping: %v", err)
	}
	var actualDatabase string
	if err := pool.QueryRow(ctx, "SELECT current_database()").Scan(&actualDatabase); err != nil {
		t.Fatalf("current_database verification: %v", err)
	}
	if actualDatabase != name {
		t.Fatalf("connected to unexpected database: got %q want %q", actualDatabase, name)
	}
	requireTestDatabase(t, actualDatabase)

	var studentCount int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM students").Scan(&studentCount); err != nil {
		t.Fatalf("student count: %v", err)
	}
	if studentCount == 0 {
		t.Skip("students table is empty; no valid cursor available")
	}

	var cursorTime time.Time
	var cursorID int
	if err := pool.QueryRow(ctx, "SELECT created_at, id FROM students ORDER BY created_at DESC, id DESC LIMIT 1").Scan(&cursorTime, &cursorID); err != nil {
		t.Fatalf("student cursor: %v", err)
	}

	studentCases := []struct {
		name string
		q    model.ListQuery
	}{
		{"student first page keyset", model.ListQuery{Limit: 2}},
		{"student cursor without filter", model.ListQuery{Limit: 5}},
		{"student cursor with search", model.ListQuery{Limit: 5, Search: "%"}},
		{"student cursor with is_active", model.ListQuery{Limit: 5, IsActive: boolPtr(true)}},
		{"student cursor with search and is_active", model.ListQuery{Limit: 5, Search: "%", IsActive: boolPtr(true)}},
	}
	for _, tc := range studentCases {
		t.Run(tc.name, func(t *testing.T) {
			query, args := buildStudentCursorQuery(tc.q, cursorTime, cursorID)
			rows, err := pool.Query(ctx, query, args...)
			if err != nil {
				t.Fatalf("query failed: %v; sql=%s; arg_types=%s", err, query, argumentTypes(args))
			}
			defer rows.Close()
			var previousTime time.Time
			previousID := 0
			seen := map[int]bool{}
			count := 0
			for rows.Next() {
				var id, nim int
				var name string
				var grade float64
				var isActive bool
				var ownerID *int
				var createdAt time.Time
				if err := rows.Scan(&id, &nim, &name, &grade, &isActive, &ownerID, &createdAt); err != nil {
					t.Fatalf("scan failed: %v", err)
				}
				if seen[id] {
					t.Fatalf("duplicate id %d", id)
				}
				if count > 0 && (createdAt.After(previousTime) || (createdAt.Equal(previousTime) && id >= previousID)) {
					t.Fatalf("ordering violation: previous=%s/%d current=%s/%d", previousTime, previousID, createdAt, id)
				}
				seen[id] = true
				previousTime, previousID = createdAt, id
				count++
			}
			if err := rows.Err(); err != nil {
				t.Fatalf("rows failed: %v", err)
			}
			t.Logf("sql=%s args=%s rows=%d", query, argumentTypes(args), count)
		})
	}

	var achievementCount int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM achievements").Scan(&achievementCount); err != nil {
		t.Fatalf("achievement count: %v", err)
	}
	if achievementCount == 0 {
		var items []model.Achievement
		var total int
		repo := NewAchievementRepository(pool)
		items, total, err = repo.FindAll(ctx, model.ListQueryAchievement{Page: 1, Limit: 5})
		if err != nil || len(items) != 0 || total != 0 {
			t.Fatalf("empty achievements query failed: err=%v items=%d total=%d", err, len(items), total)
		}
		for _, query := range []model.ListQueryAchievement{{Limit: 5}, {Limit: 5, Search: "%"}} {
			sql, args := buildAchievementCursorQuery(query, time.Now().UTC(), 1)
			rows, queryErr := pool.Query(ctx, sql, args...)
			if queryErr != nil {
				t.Fatalf("synthetic achievement cursor query failed: %v; sql=%s; arg_types=%s", queryErr, sql, argumentTypes(args))
			}
			rows.Close()
		}
		t.Log("achievement cursor SQL PASS with synthetic cursor; pagination NOT VERIFIED because table is empty")
	} else {
		t.Log("achievements contain data; cursor scenarios require a separate valid-cursor run")
	}

	cursorQuery, cursorArgs := buildStudentCursorQuery(model.ListQuery{Limit: 5}, cursorTime, cursorID)
	plans := []struct {
		name  string
		query string
		args  []any
	}{
		{name: "students first page", query: "SELECT id, created_at FROM students ORDER BY created_at DESC, id DESC LIMIT 5", args: nil},
		{name: "students cursor page", query: cursorQuery, args: cursorArgs},
	}
	for _, plan := range plans {
		t.Run("explain "+plan.name, func(t *testing.T) {
			query := plan.query
			args := plan.args
			if len(args) == 0 {
				args = nil
			}
			rows, err := pool.Query(ctx, "EXPLAIN (ANALYZE, BUFFERS) "+query, args...)
			if err != nil {
				t.Fatalf("explain failed: %v", err)
			}
			defer rows.Close()
			var lines []string
			for rows.Next() {
				var line string
				if err := rows.Scan(&line); err != nil {
					t.Fatalf("explain scan failed: %v", err)
				}
				lines = append(lines, line)
			}
			if err := rows.Err(); err != nil {
				t.Fatalf("explain rows failed: %v", err)
			}
			t.Log("query=" + query + "\n" + strings.Join(lines, "\n"))
		})
	}
}

func getenvOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func buildPostgresDSN(user, password, host, port, name, sslmode string) string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   fmt.Sprintf("%s:%s", host, port),
		Path:   name,
	}
	q := u.Query()
	q.Set("sslmode", sslmode)
	u.RawQuery = q.Encode()
	return u.String()
}

func requireTestDatabase(t *testing.T, name string) {
	t.Helper()
	normalized := strings.ToLower(strings.TrimSpace(name))
	if normalized == "" {
		t.Fatal("DB_NAME must be set explicitly for integration tests")
	}
	if normalized == "praktikum_backend" || normalized == "postgres" || normalized == "template0" || normalized == "template1" {
		t.Fatalf("refusing to run database integration test against non-test database %q", name)
	}
	if !strings.Contains(normalized, "test") {
		t.Fatalf("refusing to run database integration test against %q; DB_NAME must identify a separate test database", name)
	}
}

func argumentTypes(args []any) string {
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = fmt.Sprintf("$%d:%T", i+1, arg)
	}
	return strings.Join(parts, ", ")
}
