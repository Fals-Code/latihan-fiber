package handler

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strconv"
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
	details    map[int64]model.StudentDetail
	owners     map[int64]int64
	onDelete   func(int64)
	updated    model.StudentUpdate
	deleted    int64
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

func (f *fakeStudentStore) GetDetail(_ context.Context, id int64) (model.StudentDetail, error) {
	if f.deletedIDs[id] {
		return model.StudentDetail{}, repository.ErrNotFound
	}
	detail, ok := f.details[id]
	if !ok {
		return model.StudentDetail{}, repository.ErrNotFound
	}
	return detail, nil
}
func (f *fakeStudentStore) StudentIDByUserID(_ context.Context, userID int64) (int64, error) {
	id, ok := f.owners[userID]
	if !ok {
		return 0, repository.ErrNotFound
	}
	return id, nil
}
func (f *fakeStudentStore) Update(_ context.Context, id int64, update model.StudentUpdate) (model.Student, error) {
	if f.deletedIDs[id] {
		return model.Student{}, repository.ErrNotFound
	}
	detail, ok := f.details[id]
	if !ok {
		return model.Student{}, repository.ErrNotFound
	}
	student := detail.Student
	f.updated = update
	if update.Nama != nil {
		student.Nama = *update.Nama
	}
	if update.Prodi != nil {
		student.Prodi = *update.Prodi
	}
	if update.Angkatan != nil {
		student.Angkatan = *update.Angkatan
	}
	if update.IPKTerakhir != nil {
		student.IPKTerakhir = *update.IPKTerakhir
	}
	detail.Student = student
	f.details[id] = detail
	return student, nil
}
func (f *fakeStudentStore) SoftDelete(_ context.Context, id int64) error {
	if f.deletedIDs[id] {
		return repository.ErrNotFound
	}
	if _, ok := f.details[id]; !ok {
		return repository.ErrNotFound
	}
	f.deletedIDs[id] = true
	f.deleted = id
	if f.onDelete != nil {
		f.onDelete(id)
	}
	return nil
}

func newStudentAdminApp(t *testing.T, students []model.Student) (*fiber.App, *fakeStudentStore) {
	t.Helper()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("admin-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	authStore := &fakeAuthStore{
		users: map[string]model.User{
			"admin@example.com":   {ID: 1, Email: "admin@example.com", Password: string(passwordHash), Role: "admin"},
			"student@example.com": {ID: 2, Email: "student@example.com", Password: string(passwordHash), Role: "mahasiswa"},
		},
		active: map[int64]model.CurrentUser{1: {ID: 1, Email: "admin@example.com", Role: "admin"}, 2: {ID: 2, Email: "student@example.com", Role: "mahasiswa"}},
	}
	studentStore := &fakeStudentStore{students: students, deletedIDs: map[int64]bool{99: true}, details: map[int64]model.StudentDetail{}, owners: map[int64]int64{2: 5}}
	studentStore.onDelete = func(id int64) {
		if id == 5 {
			delete(authStore.active, 2)
		}
	}
	for _, student := range students {
		studentStore.details[student.ID] = model.StudentDetail{Student: student, Courses: []model.StudentCourse{}}
	}
	studentStore.details[5] = model.StudentDetail{Student: model.Student{ID: 5, NIM: "100000000005", Nama: "Pemilik", Prodi: "Informatika", Angkatan: 2022, IPKTerakhir: 3}, UserID: 2, Courses: []model.StudentCourse{}}
	studentStore.details[1] = model.StudentDetail{Student: model.Student{ID: 1, NIM: "100000000001", Nama: "AdminTarget", Prodi: "Prodi", Angkatan: 2021, IPKTerakhir: 3}, Courses: []model.StudentCourse{}}
	authHandler := NewAuthHandler(service.NewAuthService(authStore, testJWTSecret, time.Hour), authStore)
	studentHandler := NewStudentHandler(service.NewStudentService(studentStore))
	app := fiber.New()
	app.Post("/api/v1/auth/login", authHandler.Login)
	app.Get("/api/v1/auth/me", middleware.RequireAuth([]byte(testJWTSecret), authStore), authHandler.Me)
	app.Get("/api/v1/students", middleware.RequireAuth([]byte(testJWTSecret), authStore), middleware.RequireAdmin(), studentHandler.List)
	app.Post("/api/v1/students", middleware.RequireAuth([]byte(testJWTSecret), authStore), middleware.RequireAdmin(), studentHandler.Create)
	app.Get("/api/v1/students/:id", middleware.RequireAuth([]byte(testJWTSecret), authStore), studentHandler.Detail)
	app.Put("/api/v1/students/:id", middleware.RequireAuth([]byte(testJWTSecret), authStore), middleware.RequireAdmin(), studentHandler.Update)
	app.Delete("/api/v1/students/:id", middleware.RequireAuth([]byte(testJWTSecret), authStore), middleware.RequireAdmin(), studentHandler.Delete)
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
		if response := perform(t, app, request); response.Code != 401 {
			t.Fatalf("no token %s: %d", endpoint.path, response.Code)
		}
	}
	token := loginToken(t, app, "student@example.com")
	for _, endpoint := range []struct{ method, path string }{{"GET", "/api/v1/students"}, {"POST", "/api/v1/students"}} {
		request := httptest.NewRequest(endpoint.method, endpoint.path, strings.NewReader(`{}`))
		request.Header.Set("Authorization", "Bearer "+token)
		if response := perform(t, app, request); response.Code != 403 {
			t.Fatalf("student %s: %d", endpoint.path, response.Code)
		}
	}
}

func TestStudentListDefaultsFiltersSearchSortAndMeta(t *testing.T) {
	students := []model.Student{{ID: 1, NIM: "100000000001", Nama: "Budi", Prodi: "Informatika", Angkatan: 2023, IPKTerakhir: 3.2}, {ID: 2, NIM: "100000000002", Nama: "Andi", Prodi: "Sistem Informasi", Angkatan: 2022, IPKTerakhir: 3.8}, {ID: 3, NIM: "100000000003", Nama: "Citra", Prodi: "Informatika", Angkatan: 2023, IPKTerakhir: 3.5}, {ID: 99, NIM: "199999999999", Nama: "Deleted", Prodi: "Informatika", Angkatan: 2023, IPKTerakhir: 4}}
	app, _ := newStudentAdminApp(t, students)
	token := loginToken(t, app, "admin@example.com")
	get := func(path string) *httptest.ResponseRecorder {
		request := httptest.NewRequest("GET", path, nil)
		request.Header.Set("Authorization", "Bearer "+token)
		return perform(t, app, request)
	}
	response := get("/api/v1/students")
	var payload struct {
		Data []map[string]any `json:"data"`
		Meta struct {
			CurrentPage int `json:"current_page"`
			PerPage     int `json:"per_page"`
			Total       int `json:"total"`
			LastPage    int `json:"last_page"`
		} `json:"meta"`
	}
	if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &payload) != nil || payload.Meta.CurrentPage != 1 || payload.Meta.PerPage != 10 || payload.Meta.Total != 3 || payload.Meta.LastPage != 1 || strings.Contains(response.Body.String(), "Deleted") {
		t.Fatalf("default list: %d %s", response.Code, response.Body)
	}
	for _, test := range []struct {
		q     string
		total int
		first string
	}{{"?prodi=Informatika", 2, "Budi"}, {"?angkatan=2023", 2, "Budi"}, {"?search=100000000001", 1, "Budi"}, {"?search=andi", 1, "Andi"}, {"?sort=nama", 3, "Andi"}, {"?sort=-ipk_terakhir", 3, "Andi"}} {
		response = get("/api/v1/students" + test.q)
		var result struct {
			Data []model.Student `json:"data"`
			Meta struct {
				Total int `json:"total"`
			} `json:"meta"`
		}
		_ = json.Unmarshal(response.Body.Bytes(), &result)
		if response.Code != 200 || result.Meta.Total != test.total || len(result.Data) == 0 || result.Data[0].Nama != test.first {
			t.Fatalf("%s: %d %s", test.q, response.Code, response.Body)
		}
	}
	response = get("/api/v1/students?sort=-ipk_terakhir")
	var sorted struct {
		Data []model.Student `json:"data"`
	}
	_ = json.Unmarshal(response.Body.Bytes(), &sorted)
	if sorted.Data[0].IPKTerakhir != 3.8 {
		t.Fatalf("IPK sort: %+v", sorted.Data)
	}
	for _, q := range []string{"?sort=bad", "?per_page=51", "?per_page=0", "?page=nope", "?page=0"} {
		if response := get("/api/v1/students" + q); response.Code != 422 {
			t.Fatalf("%s: %d", q, response.Code)
		}
	}
	response = get("/api/v1/students?per_page=2&page=2")
	if response.Code != 200 || !strings.Contains(response.Body.String(), `"last_page":2`) {
		t.Fatalf("pagination: %d %s", response.Code, response.Body)
	}
	response = get("/api/v1/students?page=999")
	if response.Code != 200 || !strings.Contains(response.Body.String(), `"data":[]`) {
		t.Fatalf("out of range: %d %s", response.Code, response.Body)
	}
}

func TestStudentCreateValidationAndSafeResponse(t *testing.T) {
	app, store := newStudentAdminApp(t, nil)
	token := loginToken(t, app, "admin@example.com")
	create := func(body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("POST", "/api/v1/students", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		return perform(t, app, req)
	}
	response := create(`{"nim":"123456789012","nama":"Nama Mahasiswa","email":"new@example.com","prodi":"Informatika","angkatan":2024}`)
	if response.Code != 201 || strings.Contains(response.Body.String(), "password") || strings.Contains(response.Body.String(), "hash") {
		t.Fatalf("create: %d %s", response.Code, response.Body)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(store.created.Password), []byte("123456789012")); err != nil || store.created.IPKTerakhir != 0 {
		t.Fatalf("initial password/default IPK invalid: %v %+v", err, store.created)
	}
	for _, body := range []string{`{"nim":"","nama":"x","email":"a@b.com","prodi":"x","angkatan":2024}`, `{"nim":"123","nama":"x","email":"a@b.com","prodi":"x","angkatan":2024}`, `{"nim":"123456789012","nama":" ","email":"a@b.com","prodi":"x","angkatan":2024}`, `{"nim":"123456789012","nama":"x","email":"","prodi":"x","angkatan":2024}`, `{"nim":"123456789012","nama":"x","email":"bad","prodi":"x","angkatan":2024}`, `{"nim":"123456789012","nama":"x","email":"a@b.com","prodi":" ","angkatan":2024}`, `{"nim":"123456789012","nama":"x","email":"a@b.com","prodi":"x"}`, `{"nim":"123456789012","nama":"x","email":"a@b.com","prodi":"x","angkatan":"abc"}`, `{"nim":"123456789012","nama":"x","email":"a@b.com","prodi":"x","angkatan":9999}`, `{"nim":"123456789012","nama":"x","email":"a@b.com","prodi":"x","angkatan":2024,"ipk_terakhir":-0.1}`, `{"nim":"123456789012","nama":"x","email":"a@b.com","prodi":"x","angkatan":2024,"ipk_terakhir":4.1}`} {
		if response := create(body); response.Code != 422 {
			t.Fatalf("validation: %d %s", response.Code, response.Body)
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
		req := httptest.NewRequest("POST", "/api/v1/students", strings.NewReader(`{"nim":"123456789012","nama":"Nama","email":"a@example.com","prodi":"Prodi","angkatan":2024}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		response := perform(t, app, req)
		if response.Code != 422 || !strings.Contains(response.Body.String(), `"`+test.field+`"`) {
			t.Fatalf("duplicate: %d %s", response.Code, response.Body)
		}
	}
	app, store := newStudentAdminApp(t, nil)
	store.err = context.DeadlineExceeded
	token := loginToken(t, app, "admin@example.com")
	req := httptest.NewRequest("POST", "/api/v1/students", strings.NewReader(`{"nim":"123456789012","nama":"Nama","email":"a@example.com","prodi":"Prodi","angkatan":2024}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	response := perform(t, app, req)
	if response.Code != 500 || strings.Contains(response.Body.String(), "DeadlineExceeded") {
		t.Fatalf("internal error: %d %s", response.Code, response.Body)
	}
}

func TestStudentServiceYearUsesCurrentSystemYear(t *testing.T) {
	svc := service.NewStudentService(&fakeStudentStore{})
	input := model.NewStudent{NIM: "123456789012", Nama: "Nama", Email: "a@example.com", Prodi: "Prodi", Angkatan: time.Now().Year() + 1}
	_, validation, err := svc.Create(context.Background(), input)
	if err != nil || validation == nil || validation["angkatan"] == nil {
		t.Fatalf("dynamic year validation: %v %v", validation, err)
	}
}
func TestStudentTokenRoleIsNotTrustedWithoutDatabaseIdentity(t *testing.T) {
	app, _ := newStudentAdminApp(t, nil)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "2", "role": "admin", "exp": time.Now().Add(time.Hour).Unix()})
	raw, _ := token.SignedString([]byte(testJWTSecret))
	req := httptest.NewRequest("GET", "/api/v1/students", nil)
	req.Header.Set("Authorization", "Bearer "+raw)
	if response := perform(t, app, req); response.Code != 401 {
		t.Fatalf("mismatched role: %d", response.Code)
	}
}

func TestStudentDetailAuthorizationOwnershipCoursesAndSKSLimits(t *testing.T) {
	app, store := newStudentAdminApp(t, nil)
	store.details[5] = model.StudentDetail{Student: model.Student{ID: 5, NIM: "100000000005", Nama: "Pemilik", Prodi: "Informatika", Angkatan: 2022, IPKTerakhir: 3}, UserID: 2, Courses: []model.StudentCourse{{EnrollmentID: 8, CourseID: 4, KodeMK: "IF01", NamaMK: "Basis Data", SKS: 3, Semester: 2, TahunAkademik: "2025/2026-Ganjil"}, {EnrollmentID: 9, CourseID: 6, KodeMK: "IF02", NamaMK: "Algoritma", SKS: 4, Semester: 1, TahunAkademik: "2025/2026-Ganjil"}}, TotalSKS: 999}
	admin := loginToken(t, app, "admin@example.com")
	student := loginToken(t, app, "student@example.com")
	get := func(token, id string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("GET", "/api/v1/students/"+id, nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		return perform(t, app, req)
	}
	if get("", "5").Code != 401 {
		t.Fatal("unauthorized detail was accepted")
	}
	response := get(student, "5")
	if response.Code != 200 {
		t.Fatalf("owner with mismatched IDs: %d %s", response.Code, response.Body)
	}
	for _, want := range []string{`"total_sks":7`, `"batas_sks":24`, `"enrollment_id":8`, `"enrollment_id":9`} {
		if !strings.Contains(response.Body.String(), want) {
			t.Fatalf("missing %s: %s", want, response.Body)
		}
	}
	if get(student, "1").Code != 403 || get(admin, "1").Code != 200 || get(admin, "404").Code != 404 || get(admin, "99").Code != 404 {
		t.Fatal("detail authorization/not-found behavior mismatch")
	}
	for _, id := range []string{"abc", "0", "-1"} {
		if get(admin, id).Code != 422 {
			t.Fatalf("invalid id %s", id)
		}
	}
	emptyCourses := get(admin, "1")
	if !strings.Contains(emptyCourses.Body.String(), `"courses":[]`) || !strings.Contains(emptyCourses.Body.String(), `"total_sks":0`) {
		t.Fatalf("empty courses response: %s", emptyCourses.Body)
	}
	for _, test := range []struct {
		ipk   float64
		limit int
	}{{3, 24}, {2.99, 21}, {2.5, 21}, {2.49, 18}} {
		store.details[5] = model.StudentDetail{Student: model.Student{ID: 5, IPKTerakhir: test.ipk}, UserID: 2, Courses: []model.StudentCourse{}}
		response = get(admin, "5")
		if response.Code != 200 || !strings.Contains(response.Body.String(), `"batas_sks":`+strconv.Itoa(test.limit)) {
			t.Fatalf("IPK %.2f: %d %s", test.ipk, response.Code, response.Body)
		}
	}
}

func TestStudentUpdateValidationAuthorizationAndRestrictions(t *testing.T) {
	app, store := newStudentAdminApp(t, nil)
	admin := loginToken(t, app, "admin@example.com")
	student := loginToken(t, app, "student@example.com")
	put := func(token, id, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("PUT", "/api/v1/students/"+id, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		return perform(t, app, req)
	}
	if put("", "5", `{}`).Code != 401 || put(student, "5", `{}`).Code != 403 {
		t.Fatal("PUT auth status mismatch")
	}
	response := put(admin, "5", `{"nama":"Baru","prodi":"Sistem Informasi","angkatan":2025,"ipk_terakhir":2.5}`)
	if response.Code != 200 || strings.Contains(response.Body.String(), "password") {
		t.Fatalf("PUT response: %d %s", response.Code, response.Body)
	}
	if store.updated.Nama == nil || store.updated.Prodi == nil || store.updated.Angkatan == nil || store.updated.IPKTerakhir == nil {
		t.Fatal("allowed update fields not passed")
	}
	for _, body := range []string{`{"nim":"999999999999"}`, `{}`, `{"nama":" "}`, `{"prodi":" "}`, `{"angkatan":9999}`, `{"ipk_terakhir":-0.1}`, `{"ipk_terakhir":4.1}`, `{"angkatan":` + strconv.Itoa(time.Now().Year()+1) + `}`} {
		if response := put(admin, "5", body); response.Code != 422 {
			t.Fatalf("invalid PUT: %d %s", response.Code, response.Body)
		}
	}
	for _, id := range []string{"abc", "0", "-1"} {
		if put(admin, id, `{"nama":"Name"}`).Code != 422 {
			t.Fatalf("invalid PUT id %s", id)
		}
	}
	if put(admin, "404", `{"nama":"Name"}`).Code != 404 || put(admin, "99", `{"nama":"Name"}`).Code != 404 {
		t.Fatal("missing/deleted PUT target was accepted")
	}
}

func TestStudentSoftDeleteAndListExclusion(t *testing.T) {
	app, store := newStudentAdminApp(t, []model.Student{{ID: 5, NIM: "100000000005", Nama: "Pemilik", Prodi: "Informatika", Angkatan: 2022, IPKTerakhir: 3}})
	admin := loginToken(t, app, "admin@example.com")
	student := loginToken(t, app, "student@example.com")
	del := func(token, id string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("DELETE", "/api/v1/students/"+id, nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		return perform(t, app, req)
	}
	if del("", "5").Code != 401 || del(student, "5").Code != 403 {
		t.Fatal("DELETE auth status mismatch")
	}
	for _, id := range []string{"abc", "0", "-1"} {
		if del(admin, id).Code != 422 {
			t.Fatalf("invalid DELETE id %s", id)
		}
	}
	response := del(admin, "5")
	if response.Code != 204 || response.Body.Len() != 0 || store.deleted != 5 || !store.deletedIDs[5] {
		t.Fatalf("soft delete failed: %d %s", response.Code, response.Body)
	}
	me := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	me.Header.Set("Authorization", "Bearer "+student)
	if perform(t, app, me).Code != 401 {
		t.Fatal("soft-deleted student auth should be inactive")
	}
	if performLogin(t, app, "student@example.com", "admin-password").Code != 401 {
		t.Fatal("soft-deleted student should no longer log in")
	}
	if del(admin, "5").Code != 404 {
		t.Fatal("second delete should return 404")
	}
	get := httptest.NewRequest("GET", "/api/v1/students/5", nil)
	get.Header.Set("Authorization", "Bearer "+admin)
	if perform(t, app, get).Code != 404 {
		t.Fatal("deleted detail should return 404")
	}
	list := httptest.NewRequest("GET", "/api/v1/students", nil)
	list.Header.Set("Authorization", "Bearer "+admin)
	if strings.Contains(perform(t, app, list).Body.String(), "100000000005") {
		t.Fatal("deleted student still listed")
	}
}
