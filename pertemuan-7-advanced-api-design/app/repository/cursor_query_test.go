package repository

import (
	"strings"
	"testing"
	"time"

	"tugas1-go/pertemuan-7-advanced-api-design/app/model"
)

func TestBuildStudentCursorQueryParameters(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	tests := []struct {
		name      string
		query     model.ListQuery
		wantQuery string
		wantArgs  []any
	}{
		{"cursor without filter", model.ListQuery{Limit: 10}, "$1, $2", []any{createdAt, 7, 11}},
		{"cursor with search", model.ListQuery{Limit: 10, Search: "ana"}, "$2, $3", []any{"%ana%", createdAt, 7, 11}},
		{"cursor with active filter", model.ListQuery{Limit: 10, IsActive: boolPtr(true)}, "$2, $3", []any{true, createdAt, 7, 11}},
		{"cursor with search and active filter", model.ListQuery{Limit: 10, Search: "ana", IsActive: boolPtr(false)}, "$3, $4", []any{"%ana%", false, createdAt, 7, 11}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, args := buildStudentCursorQuery(tt.query, createdAt, 7)
			if !strings.Contains(query, tt.wantQuery) || !strings.Contains(query, "ORDER BY created_at DESC, id DESC") {
				t.Fatalf("query placeholders/order mismatch: %s", query)
			}
			assertArgs(t, args, tt.wantArgs)
		})
	}
}

func TestBuildAchievementCursorQueryParameters(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	tests := []struct {
		name      string
		query     model.ListQueryAchievement
		wantQuery string
		wantArgs  []any
	}{
		{"cursor without filter", model.ListQueryAchievement{Limit: 10}, "$1, $2", []any{createdAt, 7, 11}},
		{"cursor with search", model.ListQueryAchievement{Limit: 10, Search: "medal"}, "$2, $3", []any{"%medal%", createdAt, 7, 11}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query, args := buildAchievementCursorQuery(tt.query, createdAt, 7)
			if !strings.Contains(query, tt.wantQuery) || !strings.Contains(query, "ORDER BY created_at DESC, id DESC") {
				t.Fatalf("query placeholders/order mismatch: %s", query)
			}
			assertArgs(t, args, tt.wantArgs)
		})
	}
}

func assertArgs(t *testing.T, got, want []any) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("argument count mismatch: got %d want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("argument %d mismatch: got %#v (%T) want %#v (%T)", i+1, got[i], got[i], want[i], want[i])
		}
	}
}

func boolPtr(value bool) *bool { return &value }
