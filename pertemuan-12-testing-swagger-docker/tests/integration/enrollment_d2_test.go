//go:build integration

package integration_test

import (
	"context"
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
	"tugas1-go/pertemuan-12-testing-swagger-docker/app/repository"
	"tugas1-go/pertemuan-12-testing-swagger-docker/app/service"
	"tugas1-go/pertemuan-12-testing-swagger-docker/config"
	"tugas1-go/pertemuan-12-testing-swagger-docker/helper"
	"tugas1-go/pertemuan-12-testing-swagger-docker/route"
)

type testAPI struct {
	app *fiber.App
	jwt *helper.JWTManager
}

func newAPI(t *testing.T) testAPI {
	t.Helper()
	raw, err := repository.NewRoleRepository(pool).LoadPermissions(context.Background())
	if err != nil {
		t.Fatalf("load permissions: %v", err)
	}
	perms := helper.NewPermissionSet(raw)
	jwt := helper.NewJWTManager("integration-test-secret-32-characters-long", "integration-test", time.Minute)
	courses := repository.NewCourseRepository(pool)
	courseService := service.NewCourseService(courses, repository.NewOutboxRepository(pool), repository.NewTxManager(pool))
	app := config.NewApp(slog.New(slog.NewTextHandler(io.Discard, nil)), route.Dependencies{
		Pool: pool, JWT: jwt, Permissions: perms, CourseService: courseService,
	})
	return testAPI{app: app, jwt: jwt}
}

func (a testAPI) request(t *testing.T, userID int, method, path, body string) *http.Response {
	t.Helper()
	token, err := a.jwt.GenerateAccessToken(model.User{ID: userID, Username: "integration-user", Role: "user"})
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.app.Test(req, -1)
	if err != nil {
		t.Fatalf("request %s %s: %v", method, path, err)
	}
	return resp
}

func count(t *testing.T, query string, args ...any) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(context.Background(), query, args...).Scan(&n); err != nil {
		t.Fatalf("count query: %v", err)
	}
	return n
}

func TestSwapEndpoint_FullDestinationKeepsSource(t *testing.T) {
	a := newAPI(t)
	from := seedCourse(t, 3, 10)
	to := seedCourse(t, 3, 1)
	studentID := seedStudent(t)
	if _, err := pool.Exec(context.Background(), `INSERT INTO enrollments(course_id,user_id) VALUES($1,$2)`, from.id, studentID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(context.Background(), `UPDATE courses SET enrolled_count=1 WHERE id=$1`, from.id); err != nil {
		t.Fatal(err)
	}
	otherID := seedStudent(t)
	if _, err := pool.Exec(context.Background(), `INSERT INTO enrollments(course_id,user_id) VALUES($1,$2)`, to.id, otherID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(context.Background(), `UPDATE courses SET enrolled_count=1 WHERE id=$1`, to.id); err != nil {
		t.Fatal(err)
	}
	beforeEvents := count(t, `SELECT count(*) FROM outbox WHERE event_type=$1 AND ((payload->>'user_id')::int=$2 OR (payload->>'user_id')::int=$3)`, model.EventEnrollmentSwapped, studentID, otherID)
	resp := a.request(t, studentID, http.MethodPost, "/api/v1/enrollments/swap", `{"from_course_id":`+strconv.Itoa(from.id)+`,"to_course_id":`+strconv.Itoa(to.id)+`}`)
	defer resp.Body.Close()
	var result struct {
		Code string `json:"code"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&result)
	if resp.StatusCode != http.StatusConflict || result.Code != "COURSE_FULL" {
		t.Fatalf("swap response = %d %q, expected 409 COURSE_FULL", resp.StatusCode, result.Code)
	}
	if count(t, `SELECT count(*) FROM enrollments WHERE course_id=$1 AND user_id=$2`, from.id, studentID) != 1 {
		t.Error("student is no longer enrolled in the source course")
	}
	if count(t, `SELECT enrolled_count FROM courses WHERE id=$1`, from.id) != 1 || count(t, `SELECT enrolled_count FROM courses WHERE id=$1`, to.id) != 1 {
		t.Error("enrollment counts changed after rejected swap")
	}
	if count(t, `SELECT count(*) FROM outbox WHERE event_type=$1 AND ((payload->>'user_id')::int=$2 OR (payload->>'user_id')::int=$3)`, model.EventEnrollmentSwapped, studentID, otherID) != beforeEvents {
		t.Error("rejected swap emitted enrollment.swapped outbox event")
	}
}

func TestDropEndpoint_DeletesEnrollmentAndWritesOutbox(t *testing.T) {
	a := newAPI(t)
	course := seedCourse(t, 3, 5)
	studentID := seedStudent(t)
	if _, err := pool.Exec(context.Background(), `INSERT INTO enrollments(course_id,user_id) VALUES($1,$2)`, course.id, studentID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(context.Background(), `UPDATE courses SET enrolled_count=1 WHERE id=$1`, course.id); err != nil {
		t.Fatal(err)
	}
	resp := a.request(t, studentID, http.MethodDelete, "/api/v1/courses/"+strconv.Itoa(course.id)+"/enroll", "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("drop response = %d, expected 204", resp.StatusCode)
	}
	if count(t, `SELECT count(*) FROM enrollments WHERE course_id=$1 AND user_id=$2`, course.id, studentID) != 0 {
		t.Error("enrollment row remains after drop")
	}
	if count(t, `SELECT enrolled_count FROM courses WHERE id=$1`, course.id) != 0 {
		t.Error("course enrolled_count was not decremented")
	}
	if count(t, `SELECT count(*) FROM outbox WHERE event_type=$1 AND (payload->>'user_id')::int=$2 AND (payload->>'course_id')::int=$3`, model.EventEnrollmentDropped, studentID, course.id) != 1 {
		t.Error("expected exactly one enrollment.dropped outbox event")
	}
}
