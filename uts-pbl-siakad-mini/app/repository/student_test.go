package repository

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"tugas1-go/uts-pbl-siakad-mini/app/model"
)

type fakeStudentTx struct {
	queryCount int
	committed  bool
	rolledBack bool
	studentErr error
	failAt     int
	existsAt   int
	queries    []string
	args       [][]any
}

func (tx *fakeStudentTx) QueryRow(_ context.Context, query string, args ...any) pgx.Row {
	tx.queries = append(tx.queries, query)
	tx.args = append(tx.args, args)
	tx.queryCount++
	if tx.queryCount == tx.failAt {
		return fakeStudentRow{err: tx.studentErr}
	}
	return fakeStudentRow{queryIndex: tx.queryCount, exists: tx.queryCount == tx.existsAt}
}
func (tx *fakeStudentTx) Commit(context.Context) error { tx.committed = true; return nil }
func (tx *fakeStudentTx) Rollback(context.Context) error {
	if !tx.committed {
		tx.rolledBack = true
	}
	return nil
}

type fakeStudentRow struct {
	err        error
	queryIndex int
	exists     bool
}

func (row fakeStudentRow) Scan(dest ...any) error {
	if row.err != nil {
		return row.err
	}
	if len(dest) == 1 {
		switch value := dest[0].(type) {
		case *int64:
			*value = 10
			return nil
		case *bool:
			*value = row.exists
			return nil
		}
	}
	if len(dest) == 6 {
		*dest[0].(*int64) = 10
		*dest[1].(*string) = "123456789012"
		*dest[2].(*string) = "Nama"
		*dest[3].(*string) = "Prodi"
		*dest[4].(*int) = 2024
		*dest[5].(*float64) = 3.5
		return nil
	}
	return errors.New("unexpected scan")
}

func TestCreateStudentTransactionRollsBackOnStudentInsertFailure(t *testing.T) {
	insertErr := errors.New("student insert failed")
	tx := &fakeStudentTx{studentErr: insertErr, failAt: 4}
	starter := studentTxStarter(func(context.Context, pgx.TxOptions) (studentTransaction, error) { return tx, nil })
	_, err := createStudentTx(context.Background(), starter, model.NewStudent{Email: "student@example.com", Password: "bcrypt-hash"}, func() (string, error) { return "bcrypt-hash", nil })
	if !errors.Is(err, insertErr) {
		t.Fatalf("expected insert error, got %v", err)
	}
	if tx.committed {
		t.Fatal("transaction committed after student insert failure")
	}
	if !tx.rolledBack {
		t.Fatal("transaction was not rolled back after student insert failure")
	}
}

func TestCreateStudentTransactionChecksDuplicatesBeforeHash(t *testing.T) {
	for _, test := range []struct {
		duplicateAt int
		expected    error
	}{{1, ErrDuplicateNIM}, {2, ErrDuplicateEmail}} {
		tx := &fakeStudentTx{existsAt: test.duplicateAt}
		starter := studentTxStarter(func(context.Context, pgx.TxOptions) (studentTransaction, error) { return tx, nil })
		hashCalled := false
		_, err := createStudentTx(context.Background(), starter, model.NewStudent{NIM: "123456789012", Email: "student@example.com"}, func() (string, error) { hashCalled = true; return "hash", nil })
		if !errors.Is(err, test.expected) || hashCalled || tx.committed || !tx.rolledBack {
			t.Fatalf("duplicate check %d: err=%v hash=%v committed=%v rolledBack=%v", test.duplicateAt, err, hashCalled, tx.committed, tx.rolledBack)
		}
	}
}

func TestCreateStudentTransactionCommitsAfterBothInserts(t *testing.T) {
	tx := &fakeStudentTx{}
	starter := studentTxStarter(func(context.Context, pgx.TxOptions) (studentTransaction, error) { return tx, nil })
	hashedAfterChecks := false
	student, err := createStudentTx(context.Background(), starter, model.NewStudent{Email: "student@example.com", NIM: "123456789012", Nama: "Nama", Prodi: "Prodi", Angkatan: 2024}, func() (string, error) { hashedAfterChecks = tx.queryCount == 2; return "bcrypt-hash", nil })
	if err != nil {
		t.Fatal(err)
	}
	if !tx.committed || tx.rolledBack {
		t.Fatalf("unexpected transaction state: committed=%v rolledBack=%v", tx.committed, tx.rolledBack)
	}
	if student.ID != 10 {
		t.Fatalf("unexpected student ID: %d", student.ID)
	}
	if len(tx.queries) != 4 || !strings.Contains(tx.queries[2], "'mahasiswa'") || !hashedAfterChecks {
		t.Fatalf("new user role was not mahasiswa: %v", tx.queries)
	}
}
