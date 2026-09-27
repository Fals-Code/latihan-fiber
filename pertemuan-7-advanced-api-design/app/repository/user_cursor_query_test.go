package repository

import (
	"strings"
	"testing"
	"time"

	"tugas1-go/pertemuan-7-advanced-api-design/app/model"
)

func TestBuildUserCursorQuery(t *testing.T) {
	cursorTime := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	query, args := buildUserCursorQuery(model.UserListQuery{Limit: 2}, cursorTime, 9)
	if !strings.Contains(query, "(created_at, id) < ($1, $2)") || !strings.Contains(query, "ORDER BY created_at DESC, id DESC") || !strings.Contains(query, "LIMIT $3") {
		t.Fatalf("unexpected query: %s", query)
	}
	if len(args) != 3 || args[0] != cursorTime || args[1] != 9 || args[2] != 3 {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestBuildUserFirstPageQueryUsesLimitPlusOne(t *testing.T) {
	query, args := buildUserFirstPageQuery(model.UserListQuery{Limit: 2})
	if !strings.Contains(query, "ORDER BY created_at DESC, id DESC") || !strings.Contains(query, "LIMIT $1") {
		t.Fatalf("unexpected query: %s", query)
	}
	if len(args) != 1 || args[0] != 3 {
		t.Fatalf("unexpected args: %#v", args)
	}
}
