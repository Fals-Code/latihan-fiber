package service_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"tugas1-go/pertemuan-12-testing-swagger-docker/app/model"
	"tugas1-go/pertemuan-12-testing-swagger-docker/app/service"
	"tugas1-go/pertemuan-12-testing-swagger-docker/config"
	"tugas1-go/pertemuan-12-testing-swagger-docker/helper"
	"tugas1-go/pertemuan-12-testing-swagger-docker/route"
)

const (
	courseBackend = 1
	courseOther   = 2
	studentA      = 7
	studentB      = 8
)

type testEnv struct {
	app *fiber.App
	jwt *helper.JWTManager
}

func newTestEnv(t *testing.T) testEnv {
	t.Helper()
	reviews := &fakeReviewRepo{}
	courses := &fakeCourseRepo{
		courses: map[int]model.Course{
			courseBackend: {ID: courseBackend, Code: "IF101", Name: "Pemrograman Backend Lanjut"},
			courseOther:   {ID: courseOther, Code: "IF102", Name: "Basis Data Lanjut"},
		},
		enrolled: map[[2]int]bool{{courseBackend, studentA}: true, {courseBackend, studentB}: true},
	}
	jwtManager := helper.NewJWTManager("secret-khusus-test-minimal-32-karakter", "test", time.Minute)
	perms := helper.NewPermissionSet(map[string][]string{"admin": {"review:moderate"}})
	app := config.NewApp(slog.New(slog.NewTextHandler(io.Discard, nil)), route.Dependencies{
		JWT: jwtManager, Permissions: perms,
		ReviewService: service.NewReviewService(reviews, courses, perms),
	})
	return testEnv{app: app, jwt: jwtManager}
}

func (e testEnv) do(t *testing.T, userID int, method, path, body string) (int, map[string]any) {
	t.Helper()
	token, err := e.jwt.GenerateAccessToken(model.User{ID: userID, Username: "mhs" + strconv.Itoa(userID), Role: "user"})
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := e.app.Test(req, -1)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	var parsed map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		t.Fatalf("%s %s: response is not JSON: %v", method, path, err)
	}
	return resp.StatusCode, parsed
}

const validReview = `{"rating":5,"comment":"Materinya runtut dan contohnya nyata.","tags":["Seru","seru"]}`

func TestListReviews_NewCourseIsEmpty(t *testing.T) {
	env := newTestEnv(t)
	status, body := env.do(t, studentA, http.MethodGet, "/api/v1/courses/1/reviews", "")
	if status != http.StatusOK {
		t.Fatalf("status = %d, want 200", status)
	}
	data, ok := body["data"].([]any)
	if !ok || len(data) != 0 {
		t.Errorf("data = %v, want empty array", body["data"])
	}
}

func TestCreateReview_Success(t *testing.T) {
	env := newTestEnv(t)
	status, body := env.do(t, studentA, http.MethodPost, "/api/v1/courses/1/reviews", validReview)
	if status != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body: %v)", status, body)
	}
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("data = %v, want object", body["data"])
	}
	if tags, ok := data["tags"].([]any); !ok || len(tags) != 1 || tags[0] != "seru" {
		t.Errorf("tags = %v, want [seru]", data["tags"])
	}
	_, list := env.do(t, studentA, http.MethodGet, "/api/v1/courses/1/reviews", "")
	if n := len(list["data"].([]any)); n != 1 {
		t.Errorf("number of reviews = %d, want 1", n)
	}
}

func TestCreateReview_Rejected(t *testing.T) {
	tests := []struct {
		name, path, body string
		wantStatus       int
		wantCode         string
	}{
		{"not enrolled", "/api/v1/courses/2/reviews", validReview, http.StatusForbidden, "NOT_ENROLLED"},
		{"course does not exist", "/api/v1/courses/99/reviews", validReview, http.StatusNotFound, "NOT_FOUND"},
		{"rating outside range", "/api/v1/courses/1/reviews", `{"rating":9,"comment":"Komentar yang cukup panjang."}`, http.StatusUnprocessableEntity, "VALIDATION_ERROR"},
		{"invalid JSON", "/api/v1/courses/1/reviews", `{rating:`, http.StatusBadRequest, "BAD_REQUEST"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, body := newTestEnv(t).do(t, studentA, http.MethodPost, tt.path, tt.body)
			if status != tt.wantStatus || body["code"] != tt.wantCode {
				t.Errorf("got %d %v, want %d %s", status, body["code"], tt.wantStatus, tt.wantCode)
			}
		})
	}
}

func TestCreateReview_Twice(t *testing.T) {
	env := newTestEnv(t)
	if status, _ := env.do(t, studentB, http.MethodPost, "/api/v1/courses/1/reviews", validReview); status != http.StatusCreated {
		t.Fatalf("first review status = %d, want 201", status)
	}
	status, body := env.do(t, studentB, http.MethodPost, "/api/v1/courses/1/reviews", validReview)
	if status != http.StatusConflict || body["code"] != "ALREADY_REVIEWED" {
		t.Errorf("second review got %d %v, want 409 ALREADY_REVIEWED", status, body["code"])
	}
}
