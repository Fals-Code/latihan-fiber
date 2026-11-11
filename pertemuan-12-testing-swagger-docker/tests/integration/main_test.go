//go:build integration

package integration_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var pool *pgxpool.Pool

func TestMain(m *testing.M) {
	// Loading a local file is optional; a caller may supply TEST_DATABASE_URL
	// through the environment. Missing credentials must never report a PASS.
	_ = godotenv.Load(filepath.Join("..", "..", ".env.test"))
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "TEST_DATABASE_URL missing: integration tests NOT RUN (FAIL)")
		os.Exit(1)
	}
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, "invalid TEST_DATABASE_URL:", err)
		os.Exit(1)
	}
	if !strings.HasSuffix(cfg.ConnConfig.Database, "_test") {
		fmt.Fprintf(os.Stderr, "refusing database %q: name must end in _test\n", cfg.ConnConfig.Database)
		os.Exit(1)
	}
	// The concurrency test launches 10 requests; a one-connection pool
	// serializes them and could hide the missing-FOR-UPDATE mutation.
	cfg.MaxConns = 16
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	pool, err = pgxpool.NewWithConfig(ctx, cfg)
	if err == nil {
		err = pool.Ping(ctx)
	}
	if err == nil {
		err = applyMigrations(ctx)
	}
	cancel()
	if err != nil {
		fmt.Fprintln(os.Stderr, "integration database setup failed:", err)
		if pool != nil {
			pool.Close()
		}
		os.Exit(1)
	}
	code := m.Run()
	pool.Close()
	os.Exit(code)
}

func applyMigrations(ctx context.Context) error {
	// Bootstrap tables must exist before migrations/005_*.sql executes.
	// Sorting all matching paths together accidentally puts root migrations
	// before the bootstrap directory. Process each ordered group separately.
	for _, pattern := range []string{
		filepath.Join("..", "..", "migrations", "bootstrap", "*.sql"),
		filepath.Join("..", "..", "migrations", "*.sql"),
	} {
		files, err := filepath.Glob(pattern)
		if err != nil {
			return err
		}
		slices.Sort(files)
		for _, file := range files {
			sql, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			if _, err := pool.Exec(ctx, string(sql)); err != nil {
				return fmt.Errorf("%s: %w", filepath.Base(file), err)
			}
		}
	}
	return nil
}

// Use a random suffix rather than a timestamp to avoid collisions when
// concurrently seeding many rows in tests.
func uniqueSuffix(t *testing.T) string {
	t.Helper()
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		t.Fatalf("random test suffix: %v", err)
	}
	return hex.EncodeToString(b[:])
}

type seededCourse struct {
	id int
}

func seedCourse(t *testing.T, credits, quota int) seededCourse {
	t.Helper()
	code := "IT" + uniqueSuffix(t)
	var id int
	if err := pool.QueryRow(context.Background(), `INSERT INTO courses (code,name,credits,quota) VALUES ($1,'Integration test',$2,$3) RETURNING id`, code, credits, quota).Scan(&id); err != nil {
		t.Fatalf("seed course: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM courses WHERE id=$1`, id) })
	return seededCourse{id: id}
}

func seedStudent(t *testing.T) int {
	t.Helper()
	username := "it_" + uniqueSuffix(t)
	var id int
	if err := pool.QueryRow(context.Background(), `INSERT INTO users (username,email,password_hash,role) VALUES ($1,$2,'!','user') RETURNING id`, username, username+"@student.example").Scan(&id); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(), `DELETE FROM outbox WHERE (payload->>'user_id')::int=$1`, id)
		_, _ = pool.Exec(context.Background(), `DELETE FROM users WHERE id=$1`, id)
	})
	return id
}
