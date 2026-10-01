package seeders

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// Seed inserts the initial admin, student accounts, and course catalog.
// It is idempotent for rows identified by email, NIM, and course code.
func Seed(ctx context.Context, pool *pgxpool.Pool) error {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("Admin123!"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO users (email, password, role) VALUES ($1, $2, 'admin') ON CONFLICT (email) DO NOTHING`,
		"admin@siakad.local", string(passwordHash)); err != nil {
		return fmt.Errorf("seed admin: %w", err)
	}

	students := []struct {
		nim, nama, prodi string
		angkatan         int
		ipk              float64
	}{
		{"202200000001", "Alya Putri", "Informatika", 2022, 3.45},
		{"202200000002", "Bima Pratama", "Sistem Informasi", 2022, 3.20},
		{"202200000003", "Citra Lestari", "Informatika", 2022, 3.75},
		{"202200000004", "Dimas Saputra", "Teknik Komputer", 2022, 2.90},
		{"202200000005", "Eka Wulandari", "Sistem Informasi", 2022, 3.10},
		{"202300000001", "Fajar Nugroho", "Informatika", 2023, 3.30},
		{"202300000002", "Gita Maharani", "Sistem Informasi", 2023, 3.80},
		{"202300000003", "Hadi Kurniawan", "Teknik Komputer", 2023, 2.75},
		{"202300000004", "Intan Permata", "Informatika", 2023, 3.50},
		{"202300000005", "Joko Santoso", "Sistem Informasi", 2023, 2.40},
		{"202400000001", "Kirana Dewi", "Informatika", 2024, 3.25},
		{"202400000002", "Lukman Hakim", "Teknik Komputer", 2024, 3.05},
		{"202400000003", "Maya Anggraini", "Sistem Informasi", 2024, 3.95},
		{"202400000004", "Nanda Wijaya", "Informatika", 2024, 2.85},
		{"202400000005", "Oki Ramadhan", "Teknik Komputer", 2024, 2.20},
		{"202500000001", "Putri Amelia", "Sistem Informasi", 2025, 3.60},
		{"202500000002", "Qori Azzahra", "Informatika", 2025, 3.15},
		{"202500000003", "Rafi Firmansyah", "Teknik Komputer", 2025, 2.95},
		{"202500000004", "Salsa Nabila", "Sistem Informasi", 2025, 3.35},
		{"202500000005", "Tio Prakoso", "Informatika", 2025, 2.65},
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin student seed: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, student := range students {
		hash, err := bcrypt.GenerateFromPassword([]byte(student.nim), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("hash password for %s: %w", student.nim, err)
		}
		email := student.nim + "@student.siakad.local"
		if _, err := tx.Exec(ctx,
			`INSERT INTO users (email, password, role) VALUES ($1, $2, 'mahasiswa') ON CONFLICT (email) DO NOTHING`,
			email, string(hash)); err != nil {
			return fmt.Errorf("seed user for %s: %w", student.nim, err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO students (user_id, nim, nama, prodi, angkatan, ipk_terakhir)
			 SELECT id, $1, $2, $3, $4, $5 FROM users WHERE email = $6
			 ON CONFLICT (nim) DO NOTHING`,
			student.nim, student.nama, student.prodi, student.angkatan, student.ipk, email); err != nil {
			return fmt.Errorf("seed student %s: %w", student.nim, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit student seed: %w", err)
	}

	courses := []struct {
		code, name       string
		sks, term, quota int
	}{
		{"IF101", "Algoritma dan Pemrograman", 3, 1, 40},
		{"IF102", "Struktur Data", 3, 2, 35},
		{"IF201", "Basis Data", 3, 3, 35},
		{"IF202", "Pemrograman Web", 3, 4, 30},
		{"IF301", "Rekayasa Perangkat Lunak", 3, 5, 30},
		{"SI101", "Pengantar Sistem Informasi", 2, 1, 40},
		{"SI201", "Analisis Proses Bisnis", 3, 3, 30},
		{"TK101", "Pengantar Jaringan Komputer", 3, 1, 35},
		{"TK202", "Sistem Operasi", 3, 4, 30},
		{"MKU101", "Pancasila", 2, 1, 50},
	}
	for _, course := range courses {
		if _, err := pool.Exec(ctx,
			`INSERT INTO courses (kode_mk, nama_mk, sks, semester, kuota) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (kode_mk) DO NOTHING`,
			course.code, course.name, course.sks, course.term, course.quota); err != nil {
			return fmt.Errorf("seed course %s: %w", course.code, err)
		}
	}
	return nil
}
