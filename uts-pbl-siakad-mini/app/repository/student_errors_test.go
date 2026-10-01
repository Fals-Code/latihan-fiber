package repository

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestDuplicateConstraintErrorsAreMapped(t *testing.T) {
	for _, test := range []struct {
		constraint string
		want       error
	}{
		{"students_nim_key", ErrDuplicateNIM}, {"users_email_key", ErrDuplicateEmail},
	} {
		got := duplicateError(&pgconn.PgError{Code: "23505", ConstraintName: test.constraint})
		if !errors.Is(got, test.want) {
			t.Fatalf("%s: expected %v, got %v", test.constraint, test.want, got)
		}
	}
}
