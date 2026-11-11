//go:build integration

package integration_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"tugas1-go/pertemuan-12-testing-swagger-docker/app/model"
	"tugas1-go/pertemuan-12-testing-swagger-docker/app/repository"
)

func TestIncrementEnrolled_StopsAtQuota(t *testing.T) {
	ctx := context.Background()
	course := seedCourse(t, 3, 2)
	repo := repository.NewCourseRepository(pool)
	for i := 0; i < 2; i++ {
		if _, err := repo.IncrementEnrolled(ctx, course.id); err != nil {
			t.Fatalf("increment %d: %v", i+1, err)
		}
	}
	if _, err := repo.IncrementEnrolled(ctx, course.id); !errors.Is(err, repository.ErrCourseFull) {
		t.Fatalf("third increment error = %v, want ErrCourseFull", err)
	}
	got, err := repo.FindByID(ctx, course.id)
	if err != nil || got.EnrolledCount != 2 {
		t.Fatalf("course = %+v, %v; want enrolled_count 2", got, err)
	}
}

func TestInsertEnrollment_Duplicate(t *testing.T) {
	ctx := context.Background()
	course, studentID := seedCourse(t, 3, 40), seedStudent(t)
	repo := repository.NewCourseRepository(pool)
	if err := repo.InsertEnrollment(ctx, course.id, studentID); err != nil {
		t.Fatalf("first enrollment: %v", err)
	}
	if err := repo.InsertEnrollment(ctx, course.id, studentID); !errors.Is(err, repository.ErrDuplicate) {
		t.Fatalf("second enrollment error = %v, want ErrDuplicate", err)
	}
}

func TestWithTx_RollsBackEveryStep(t *testing.T) {
	ctx := context.Background()
	course, studentID := seedCourse(t, 3, 40), seedStudent(t)
	courses, outbox := repository.NewCourseRepository(pool), repository.NewOutboxRepository(pool)
	stop := errors.New("intentional rollback")
	err := repository.NewTxManager(pool).WithTx(ctx, func(tx repository.DBTX) error {
		if _, err := courses.WithTx(tx).IncrementEnrolled(ctx, course.id); err != nil {
			return err
		}
		if err := outbox.WithTx(tx).Insert(ctx, model.EventEnrollmentCreated, model.EnrollmentEvent{UserID: studentID, CourseID: course.id}); err != nil {
			return err
		}
		return stop
	})
	if !errors.Is(err, stop) {
		t.Fatalf("WithTx error = %v, want intentional rollback", err)
	}
	got, err := courses.FindByID(ctx, course.id)
	if err != nil || got.EnrolledCount != 0 {
		t.Fatalf("course after rollback = %+v, %v; want enrolled_count 0", got, err)
	}
	if n := count(t, `SELECT COUNT(*) FROM outbox WHERE (payload->>'user_id')::int=$1 AND (payload->>'course_id')::int=$2`, studentID, course.id); n != 0 {
		t.Fatalf("outbox rows after rollback = %d, want 0", n)
	}
}

func TestClaim_OnlyOnce(t *testing.T) {
	ctx := context.Background()
	studentID := seedStudent(t)
	var outboxID int64
	if err := pool.QueryRow(ctx, `INSERT INTO outbox (event_type,payload) VALUES ($1,jsonb_build_object('user_id',$2::int)) RETURNING id`, model.EventEnrollmentCreated, studentID).Scan(&outboxID); err != nil {
		t.Fatalf("seed outbox: %v", err)
	}
	t.Cleanup(func() { _, _ = pool.Exec(context.Background(), `DELETE FROM outbox WHERE id=$1`, outboxID) })
	n := model.Notification{OutboxID: outboxID, UserID: studentID, Channel: "email", Subject: "test", Body: "test"}
	repo := repository.NewNotificationRepository(pool)
	first, err := repo.Claim(ctx, n)
	if err != nil || !first {
		t.Fatalf("first claim = %v, %v; want true, nil", first, err)
	}
	second, err := repo.Claim(ctx, n)
	if err != nil || second {
		t.Fatalf("second claim = %v, %v; want false, nil", second, err)
	}
}

func TestEnrollEndpoint_WritesOutboxOnlyOnSuccess(t *testing.T) {
	a := newAPI(t)
	course, studentID := seedCourse(t, 3, 40), seedStudent(t)
	resp := a.request(t, studentID, http.MethodPost, "/api/v1/courses/"+strconv.Itoa(course.id)+"/enroll", "")
	var first struct {
		Code string `json:"code"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&first)
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("first enrollment status = %d, want 201", resp.StatusCode)
	}
	resp = a.request(t, studentID, http.MethodPost, "/api/v1/courses/"+strconv.Itoa(course.id)+"/enroll", "")
	var second struct {
		Code string `json:"code"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&second)
	resp.Body.Close()
	if resp.StatusCode != http.StatusConflict || second.Code != "ALREADY_ENROLLED" {
		t.Fatalf("duplicate enrollment = %d %q, want 409 ALREADY_ENROLLED", resp.StatusCode, second.Code)
	}
	if n := count(t, `SELECT COUNT(*) FROM outbox WHERE event_type=$1 AND (payload->>'user_id')::int=$2 AND (payload->>'course_id')::int=$3`, model.EventEnrollmentCreated, studentID, course.id); n != 1 {
		t.Fatalf("enrollment outbox events = %d, want exactly 1", n)
	}
}

func TestEnrollEndpoint_ConcurrentCreditLimit(t *testing.T) {
	a := newAPI(t)
	studentID := seedStudent(t)
	const attempts = 10
	courses := make([]seededCourse, attempts)
	for i := range courses {
		courses[i] = seedCourse(t, 3, 40)
	}
	token, err := a.jwt.GenerateAccessToken(model.User{ID: studentID, Username: "integration-user", Role: "user"})
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	type result struct {
		status int
		err    error
	}
	results := make(chan result, attempts)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for _, c := range courses {
		wg.Add(1)
		go func(courseID int) {
			defer wg.Done()
			<-start
			req := httptest.NewRequest(http.MethodPost, "/api/v1/courses/"+strconv.Itoa(courseID)+"/enroll", strings.NewReader(""))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			resp, err := a.app.Test(req, -1)
			if err != nil {
				results <- result{err: err}
				return
			}
			_, readErr := io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			if readErr != nil {
				results <- result{err: readErr}
				return
			}
			results <- result{status: resp.StatusCode}
		}(c.id)
	}
	close(start)
	wg.Wait()
	close(results)
	created, conflicts := 0, 0
	for result := range results {
		if result.err != nil {
			t.Errorf("concurrent enrollment request: %v", result.err)
			continue
		}
		switch result.status {
		case http.StatusCreated:
			created++
		case http.StatusConflict:
			conflicts++
		default:
			t.Errorf("unexpected enrollment status %d", result.status)
		}
	}
	if created != 8 || conflicts != 2 {
		t.Errorf("concurrent enrollment results: created=%d conflicts=%d, want 8 and 2", created, conflicts)
	}
	if n := count(t, `SELECT COALESCE(SUM(c.credits),0) FROM enrollments e JOIN courses c ON c.id=e.course_id WHERE e.user_id=$1`, studentID); n != 24 {
		t.Errorf("enrolled credits = %d, want 24", n)
	}
}
