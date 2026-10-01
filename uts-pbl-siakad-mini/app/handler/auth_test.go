package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"tugas1-go/uts-pbl-siakad-mini/app/model"
	"tugas1-go/uts-pbl-siakad-mini/app/repository"
	"tugas1-go/uts-pbl-siakad-mini/app/service"
	"tugas1-go/uts-pbl-siakad-mini/middleware"
)

const testJWTSecret = "test-secret-that-is-long-enough-for-tests"

type fakeAuthStore struct {
	users  map[string]model.User
	active map[int64]model.CurrentUser
}

func (s *fakeAuthStore) FindUserByEmail(_ context.Context, email string) (model.User, error) {
	user, ok := s.users[email]
	if !ok {
		return model.User{}, repository.ErrNotFound
	}
	return user, nil
}

func (s *fakeAuthStore) FindActiveAccount(_ context.Context, id int64) (model.CurrentUser, error) {
	user, ok := s.active[id]
	if !ok {
		return model.CurrentUser{}, repository.ErrNotFound
	}
	return user, nil
}

func newAuthTestApp(t *testing.T) (*fiber.App, *fakeAuthStore) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	student := &model.StudentProfile{NIM: "123456789012", Nama: "Student", Prodi: "Informatika", Angkatan: 2024}
	store := &fakeAuthStore{
		users: map[string]model.User{
			"admin@example.com":   {ID: 1, Email: "admin@example.com", Password: string(hash), Role: "admin"},
			"student@example.com": {ID: 2, Email: "student@example.com", Password: string(hash), Role: "mahasiswa"},
		},
		active: map[int64]model.CurrentUser{
			1: {ID: 1, Email: "admin@example.com", Role: "admin"},
			2: {ID: 2, Email: "student@example.com", Role: "mahasiswa", Student: student},
		},
	}
	authService := service.NewAuthService(store, testJWTSecret, time.Hour)
	authHandler := NewAuthHandler(authService, store)
	app := fiber.New()
	app.Post("/api/v1/auth/login", middleware.LoginFailureLimiter(), authHandler.Login)
	app.Get("/api/v1/auth/me", middleware.RequireAuth([]byte(testJWTSecret), store), authHandler.Me)
	return app, store
}

func perform(t *testing.T, app *fiber.App, request *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	result := httptest.NewRecorder()
	result.WriteHeader(response.StatusCode)
	_, _ = io.Copy(result, response.Body)
	return result
}

func performLogin(t *testing.T, app *fiber.App, email, password string) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	request := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(string(body)))
	request.Header.Set("Content-Type", "application/json")
	return perform(t, app, request)
}

func TestLoginSuccessForAdminAndStudent(t *testing.T) {
	for _, email := range []string{"admin@example.com", "student@example.com"} {
		t.Run(email, func(t *testing.T) {
			app, _ := newAuthTestApp(t)
			response := performLogin(t, app, email, "correct-password")
			if response.Code != fiber.StatusOK {
				t.Fatalf("expected 200, got %d: %s", response.Code, response.Body)
			}
			if !strings.Contains(response.Body.String(), `"token_type":"Bearer"`) || strings.Contains(response.Body.String(), "correct-password") {
				t.Fatalf("unexpected login response: %s", response.Body)
			}
		})
	}
}

func TestLoginCredentialAndValidationFailures(t *testing.T) {
	tests := []struct {
		name, email, password string
		status                int
	}{
		{"bad password", "admin@example.com", "wrong-password", 401},
		{"unknown email", "unknown@example.com", "correct-password", 401},
		{"invalid email", "not-an-email", "correct-password", 422},
		{"short password", "admin@example.com", "short", 422},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			app, _ := newAuthTestApp(t)
			response := performLogin(t, app, test.email, test.password)
			if response.Code != test.status {
				t.Fatalf("expected %d, got %d", test.status, response.Code)
			}
		})
	}
}

func TestMeRequiresValidTokenAndReturnsAccount(t *testing.T) {
	app, _ := newAuthTestApp(t)
	for _, header := range []string{"", "Bearer", "Bearer broken", "Bearer a.b.c"} {
		request := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
		if header != "" {
			request.Header.Set("Authorization", header)
		}
		response := perform(t, app, request)
		if response.Code != 401 {
			t.Fatalf("header %q: expected 401, got %d", header, response.Code)
		}
	}

	for _, account := range []struct{ email, role string }{{"admin@example.com", "admin"}, {"student@example.com", "mahasiswa"}} {
		login := performLogin(t, app, account.email, "correct-password")
		var envelope struct {
			Data struct {
				Role    string                `json:"role"`
				Student *model.StudentProfile `json:"student"`
			} `json:"data"`
		}
		if err := json.Unmarshal(login.Body.Bytes(), &envelope); err != nil {
			t.Fatal(err)
		}
		if login.Code != 200 {
			t.Fatalf("login status %d", login.Code)
		}
		var responseBody map[string]any
		if err := json.Unmarshal(login.Body.Bytes(), &responseBody); err != nil {
			t.Fatal(err)
		}
		data := responseBody["data"].(map[string]any)
		token := data["access_token"].(string)
		request := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
		request.Header.Set("Authorization", "Bearer "+token)
		response := perform(t, app, request)
		if response.Code != 200 || !strings.Contains(response.Body.String(), `"role":"`+account.role+`"`) {
			t.Fatalf("unexpected response: %d %s", response.Code, response.Body)
		}
		if account.role == "mahasiswa" && (!strings.Contains(response.Body.String(), `"nim":"123456789012"`) || !strings.Contains(response.Body.String(), `"nama":"Student"`)) {
			t.Fatalf("student details missing: %s", response.Body)
		}
	}
}

func TestInvalidClaimTokensRejected(t *testing.T) {
	app, _ := newAuthTestApp(t)
	for _, claims := range []jwt.MapClaims{
		{"sub": "0", "role": "admin", "exp": time.Now().Add(time.Hour).Unix()},
		{"sub": "1", "role": "unknown", "exp": time.Now().Add(time.Hour).Unix()},
		{"role": "admin", "exp": time.Now().Add(time.Hour).Unix()},
	} {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		raw, err := token.SignedString([]byte(testJWTSecret))
		if err != nil {
			t.Fatal(err)
		}
		request := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
		request.Header.Set("Authorization", "Bearer "+raw)
		if response := perform(t, app, request); response.Code != 401 {
			t.Fatalf("invalid claims: expected 401, got %d", response.Code)
		}
	}
}

func TestExpiredTokenRejected(t *testing.T) {
	app, _ := newAuthTestApp(t)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "1", "role": "admin", "exp": time.Now().Add(-time.Minute).Unix()})
	raw, err := token.SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	request.Header.Set("Authorization", "Bearer "+raw)
	response := perform(t, app, request)
	if response.Code != 401 {
		t.Fatalf("expected 401, got %d", response.Code)
	}
}

func TestSoftDeletedStudentCannotLoginOrUseExistingToken(t *testing.T) {
	app, store := newAuthTestApp(t)
	login := performLogin(t, app, "student@example.com", "correct-password")
	if login.Code != 200 {
		t.Fatalf("expected initial login 200, got %d", login.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(login.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	token := body["data"].(map[string]any)["access_token"].(string)
	delete(store.active, 2)
	if response := performLogin(t, app, "student@example.com", "correct-password"); response.Code != 401 {
		t.Fatalf("deleted login: expected 401, got %d", response.Code)
	}
	request := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := perform(t, app, request)
	if response.Code != 401 {
		t.Fatalf("existing token: expected 401, got %d", response.Code)
	}
}

func TestValidationAndSuccessDoNotIncreaseFailureCount(t *testing.T) {
	app, _ := newAuthTestApp(t)
	for i := 0; i < 5; i++ {
		if response := performLogin(t, app, "admin@example.com", "wrong-password"); response.Code != 401 {
			t.Fatalf("failure %d: %d", i+1, response.Code)
		}
	}
	if response := performLogin(t, app, "bad-email", "correct-password"); response.Code != 429 {
		t.Fatalf("blocked after five failures: %d", response.Code)
	}

	app, _ = newAuthTestApp(t)
	if response := performLogin(t, app, "bad-email", "correct-password"); response.Code != 422 {
		t.Fatalf("expected 422, got %d", response.Code)
	}
	if response := performLogin(t, app, "admin@example.com", "correct-password"); response.Code != 200 {
		t.Fatalf("successful login failed: %d", response.Code)
	}
	for i := 0; i < 5; i++ {
		if response := performLogin(t, app, "admin@example.com", "wrong-password"); response.Code != 401 {
			t.Fatalf("failure %d after validation/success: %d", i+1, response.Code)
		}
	}
	if response := performLogin(t, app, "admin@example.com", "wrong-password"); response.Code != 429 {
		t.Fatalf("sixth failure: expected 429, got %d", response.Code)
	}
}
