package handler

import (
	"context"
	"encoding/json"
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

type fakeStudentStore struct {
	students   []model.Student
	deletedIDs map[int64]bool
	created    model.NewStudent
	err        error
}

func (f *fakeStudentStore) List(_ context.Context, filters model.StudentFilters) ([]model.Student, int64, error) {
	filtered := make([]model.Student, 0)
	for _, student := range f.students {
		if f.deletedIDs[student.ID] {
			continue
		}
		if filters.Prodi != "" && student.Prodi != filters.Prodi {
			continue
		}
		if filters.Angkatan != nil && student.Angkatan != *filters.Angkatan {
			continue
		}
		search := strings.ToLower(filters.Search)
		if search != "" && !strings.Contains(strings.ToLower(student.NIM), search) && !strings.Contains(strings.ToLower(student.Nama), search) {
			continue
		}
		filtered = append(filtered, student)
	}
	if filters.Sort == "nama" {
		for i := 0; i < len(filtered); i++ {
			for j := i + 1; j < len(filtered); j++ {
				if filtered[j].Nama < filtered[i].Nama {
					filtered[i], filtered[j] = filtered[j], filtered[i]
				}
			}
		}
	}
	if filters.Sort == "-ipk_terakhir" {
		for i := 0; i < len(filtered); i++ {
			for j := i + 1; j < len(filtered); j++ {
				if filtered[j].IPKTerakhir > filtered[i].IPKTerakhir {
					filtered[i], filtered[j] = filtered[j], filtered[i]
				}
			}
		}
	}
	total := int64(len(filtered))
	start := (filters.Page - 1) * filters.PerPage
	if start >= len(filtered) {
		return []model.Student{}, total, nil
	}
	end := start + filters.PerPage
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], total, nil
}

func (f *fakeStudentStore) Create(_ context.Context, student model.NewStudent, hashPassword func() (string, error)) (model.Student, error) {
	if f.err != nil {
		return model.Student{}, f.err
	}
	var err error
	student.Password, err = hashPassword()
	if err != nil {
		return model.Student{}, err
	}
	f.created = student
	return model.Student{ID: 50, NIM: student.NIM, Nama: student.Nama, Prodi: student.Prodi, Angkatan: student.Angkatan, IPKTerakhir: student.IPKTerakhir}, nil
}

func newStudentAdminApp(t *testing.T, students []model.Student) (*fiber.App, *fakeStudentStore) {
	t.Helper()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("admin-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	authStore := &fakeAuthStore{
		users:  map[string]model.User{"admin@example.com": {ID: 1, Email: "admin@example.com", Password: string(passwordHash), Role: "admin"}, "student@example.com": {ID: 2, Email: "student@example.com", Password: string(passwordHash), Role: "mahasiswa"}},
		active: map[int64]model.CurrentUser{1: {ID: 1, Email: "admin@example.com", Role: "admin"}, 2: {ID: 2, Email: "student@example.com", Role: "mahasiswa"}},
	}
	studentStore := &fakeStudentStore{students: students, deletedIDs: map[int64]bool{99: true}}
	authHandler := NewAuthHandler(service.NewAuthService(authStore, testJWTSecret, time.Hour), authStore)
	studentHandler := NewStudentHandler(service.NewStudentService(studentStore))
	app := fiber.New()
	app.Post("/api/v1/auth/login", authHandler.Login)
	app.Get("/api/v1/auth/me", middleware.RequireAuth([]byte(testJWTSecret), authStore), authHandler.Me)
	app.Get("/api/v1/students", middleware.RequireAuth([]byte(testJWTSecret), authStore), middleware.RequireAdmin(), studentHandler.List)
	app.Post("/api/v1/students", middleware.RequireAuth([]byte(testJWTSecret), authStore), middleware.RequireAdmin(), studentHandler.Create)
	return app, studentStore
}

func loginToken(t *testing.T, app *fiber.App, email string) string {
	t.Helper()
	response := performLogin(t, app, email, "admin-password")
	var payload struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	return payload.Data.AccessToken
}

func TestStudentEndpointsRequireAdmin(t *testing.T) {
	app, _ := newStudentAdminApp(t, nil)
	for _, endpoint := range []struct{ method, path string }{{"GET", "/api/v1/students"}, {"POST", "/api/v1/students"}} {
		request := httptest.NewRequest(endpoint.method, endpoint.path, strings.NewReader(`{}`))
		response := perform(t, app, request)
		if response.Code != 401 {
			t.Fatalf("no token %s: expected 401, got %d", endpoint.path, response.Code)
		}
	}
	token := loginToken(t, app, "student@example.com")
	for _, endpoint := range []struct{ method, path string }{{"GET", "/api/v1/students"}, {"POST", "/api/v1/students"}} {
		request := httptest.NewRequest(endpoint.method, endpoint.path, strings.NewReader(`{}`))
		request.Header.Set("Authorization", "Bearer "+token)
		response := perform(t, app, request)
		if response.Code != 403 {
			t.Fatalf("student %s: expected 403, got %d", endpoint.path, response.Code)
		}
	}
}

func TestStudentListDefaultsFiltersSearchSortAndMeta(t *testing.T) {
	students := []model.Student{
		{ID: 1, NIM: "100000000001", Nama: "Budi", Prodi: "Informatika", Angkatan: 2023, IPKTerakhir: 3.2},
		{ID: 2, NIM: "100000000002", Nama: "Andi", Prodi: "Sistem Informasi", Angkatan: 2022, IPKTerakhir: 3.8},
		{ID: 3, NIM: "100000000003", Nama: "Citra", Prodi: "Informatika", Angkatan: 2023, IPKTerakhir: 3.5},
		{ID: 99, NIM: "199999999999", Nama: "Deleted", Prodi: "Informatika", Angkatan: 2023, IPKTerakhir: 4.0},
	}
	app, _ := newStudentAdminApp(t, students)
	token := loginToken(t, app, "admin@example.com")
	get := func(path string) *httptest.ResponseRecorder {
		request := httptest.NewRequest("GET", path, nil)
		request.Header.Set("Authorization", "Bearer "+token)
		return perform(t, app, request)
	}
	response := get("/api/v1/students")
	if response.Code != 200 {
		t.Fatalf("default list status %d: %s", response.Code, response.Body)
	}
	var payload struct {
		Data []map[string]any `json:"data"`
		Meta struct {
			CurrentPage int `json:"current_page"`
			PerPage     int `json:"per_page"`
			Total       int `json:"total"`
			LastPage    int `json:"last_page"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Meta.CurrentPage != 1 || payload.Meta.PerPage != 10 || payload.Meta.Total != 3 || payload.Meta.LastPage != 1 {
		t.Fatalf("bad default meta: %+v", payload.Meta)
	}
	if strings.Contains(response.Body.String(), "Deleted") {
		t.Fatal("soft-deleted student was returned")
	}
	for _, test := range []struct {
		query     string
		total     int
		firstName string
	}{
		{"?prodi=Informatika", 2, "Budi"}, {"?angkatan=2023", 2, "Budi"},
		{"?search=100000000001", 1, "Budi"}, {"?search=andi", 1, "Andi"},
		{"?sort=nama", 3, "Andi"}, {"?sort=-ipk_terakhir", 3, "Andi"},
	} {
		response := get("/api/v1/students" + test.query)
		if response.Code != 200 {
			t.Fatalf("%s status %d: %s", test.query, response.Code, response.Body)
		}
		var result struct {
			Data []model.Student `json:"data"`
			Meta struct {
				Total int `json:"total"`
			} `json:"meta"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if result.Meta.Total != test.total || len(result.Data) == 0 || result.Data[0].Nama != test.firstName {
			t.Fatalf("%s returned unexpected result: %+v", test.query, result)
		}
	}
	ipkSorted := get("/api/v1/students?sort=-ipk_terakhir")
	var ipkResult struct {
		Data []model.Student `json:"data"`
	}
	if err := json.Unmarshal(ipkSorted.Body.Bytes(), &ipkResult); err != nil {
		t.Fatal(err)
	}
	if ipkResult.Data[0].IPKTerakhir != 3.8 {
		t.Fatalf("IPK sort did not descend: %+v", ipkResult.Data)
	}
	if response := get("/api/v1/students?sort=bad"); response.Code != 422 {
		t.Fatalf("invalid sort expected 422, got %d", response.Code)
	}
	if response := get("/api/v1/students?per_page=51"); response.Code != 422 {
		t.Fatalf("oversized per_page expected 422, got %d", response.Code)
	}
	if response := get("/api/v1/students?per_page=0"); response.Code != 422 {
		t.Fatalf("invalid per_page expected 422, got %d", response.Code)
	}
	if response := get("/api/v1/students?page=nope"); response.Code != 422 {
		t.Fatalf("invalid page expected 422, got %d", response.Code)
	}
	pageTwo := get("/api/v1/students?per_page=2&page=2")
	if pageTwo.Code != 200 || !strings.Contains(pageTwo.Body.String(), `"last_page":2`) || !strings.Contains(pageTwo.Body.String(), `"current_page":2`) {
		t.Fatalf("bad pagination: %d %s", pageTwo.Code, pageTwo.Body)
	}
	if response := get("/api/v1/students?page=0"); response.Code != 422 {
		t.Fatalf("invalid page expected 422, got %d", response.Code)
	}
	if response := get("/api/v1/students?page=999"); response.Code != 200 || !strings.Contains(response.Body.String(), `"data":[]`) {
		t.Fatalf("out-of-range page: %d %s", response.Code, response.Body)
	}
}

func TestStudentCreateValidationAndSafeResponse(t *testing.T) {
	app, store := newStudentAdminApp(t, nil)
	token := loginToken(t, app, "admin@example.com")
	create := func(body string) *httptest.ResponseRecorder {
		request := httptest.NewRequest("POST", "/api/v1/students", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+token)
		return perform(t, app, request)
	}
	valid := `{"nim":"123456789012","nama":"Nama Mahasiswa","email":"new@example.com","prodi":"Informatika","angkatan":2024}`
	response := create(valid)
	if response.Code != 201 {
		t.Fatalf("create status %d: %s", response.Code, response.Body)
	}
	if strings.Contains(response.Body.String(), "password") || strings.Contains(response.Body.String(), "hash") {
		t.Fatalf("sensitive data leaked: %s", response.Body)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(store.created.Password), []byte("123456789012")); err != nil {
		t.Fatal("initial password was not NIM bcrypt hash")
	}
	if store.created.IPKTerakhir != 0 {
		t.Fatalf("omitted IPK should default to schema value 0.00, got %v", store.created.IPKTerakhir)
	}
	for _, invalid := range []string{
		`{"nim":"","nama":"x","email":"a@b.com","prodi":"x","angkatan":2024}`,
		`{"nim":"123","nama":"x","email":"a@b.com","prodi":"x","angkatan":2024}`,
		`{"nim":"123456789012","nama":" ","email":"a@b.com","prodi":"x","angkatan":2024}`,
		`{"nim":"123456789012","nama":"x","email":"","prodi":"x","angkatan":2024}`,
		`{"nim":"123456789012","nama":"x","email":"bad","prodi":"x","angkatan":2024}`,
		`{"nim":"123456789012","nama":"x","email":"a@b.com","prodi":" ","angkatan":2024}`,
		`{"nim":"123456789012","nama":"x","email":"a@b.com","prodi":"x"}`,
		`{"nim":"123456789012","nama":"x","email":"a@b.com","prodi":"x","angkatan":"abc"}`,
		`{"nim":"123456789012","nama":"x","email":"a@b.com","prodi":"x","angkatan":9999}`,
		`{"nim":"123456789012","nama":"x","email":"a@b.com","prodi":"x","angkatan":2024,"ipk_terakhir":-0.1}`,
		`{"nim":"123456789012","nama":"x","email":"a@b.com","prodi":"x","angkatan":2024,"ipk_terakhir":4.1}`,
	} {
		if response := create(invalid); response.Code != 422 {
			t.Fatalf("invalid payload expected 422, got %d: %s", response.Code, response.Body)
		}
	}
}

func TestStudentCreateDuplicateErrorsAndInternalSafety(t *testing.T) {
	for _, test := range []struct {
		err   error
		field string
	}{{repository.ErrDuplicateNIM, "nim"}, {repository.ErrDuplicateEmail, "email"}} {
		app, store := newStudentAdminApp(t, nil)
		store.err = test.err
		token := loginToken(t, app, "admin@example.com")
		request := httptest.NewRequest("POST", "/api/v1/students", strings.NewReader(`{"nim":"123456789012","nama":"Nama","email":"a@example.com","prodi":"Prodi","angkatan":2024}`))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+token)
		response := perform(t, app, request)
		if response.Code != 422 || !strings.Contains(response.Body.String(), `"`+test.field+`"`) {
			t.Fatalf("duplicate %s response: %d %s", test.field, response.Code, response.Body)
		}
	}
	app, store := newStudentAdminApp(t, nil)
	store.err = context.DeadlineExceeded
	token := loginToken(t, app, "admin@example.com")
	request := httptest.NewRequest("POST", "/api/v1/students", strings.NewReader(`{"nim":"123456789012","nama":"Nama","email":"a@example.com","prodi":"Prodi","angkatan":2024}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	response := perform(t, app, request)
	if response.Code != 500 || strings.Contains(response.Body.String(), "DeadlineExceeded") {
		t.Fatalf("internal error response: %d %s", response.Code, response.Body)
	}
}

func TestStudentServiceYearUsesCurrentSystemYear(t *testing.T) {
	store := &fakeStudentStore{}
	svc := service.NewStudentService(store)
	input := model.NewStudent{NIM: "123456789012", Nama: "Nama", Email: "a@example.com", Prodi: "Prodi", Angkatan: time.Now().Year() + 1}
	_, validation, err := svc.Create(context.Background(), input)
	if err != nil || validation == nil || validation["angkatan"] == nil {
		t.Fatalf("expected dynamic year validation: %v %v", validation, err)
	}
}

func TestStudentTokenRoleIsNotTrustedWithoutDatabaseIdentity(t *testing.T) {
	app, _ := newStudentAdminApp(t, nil)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "2", "role": "admin", "exp": time.Now().Add(time.Hour).Unix()})
	raw, err := token.SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest("GET", "/api/v1/students", nil)
	request.Header.Set("Authorization", "Bearer "+raw)
	if response := perform(t, app, request); response.Code != 401 {
		t.Fatalf("mismatched identity should be rejected, got %d", response.Code)
	}
}
