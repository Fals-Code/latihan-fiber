//go:build integration

package route_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"tugas1-go/pertemuan-10-api-mongodb/app/model"
	p10repository "tugas1-go/pertemuan-10-api-mongodb/app/repository"
	p10service "tugas1-go/pertemuan-10-api-mongodb/app/service"
	p10config "tugas1-go/pertemuan-10-api-mongodb/config"
	p10database "tugas1-go/pertemuan-10-api-mongodb/database"
	p10route "tugas1-go/pertemuan-10-api-mongodb/route"
	p7model "tugas1-go/pertemuan-7-advanced-api-design/app/model"
	p7repository "tugas1-go/pertemuan-7-advanced-api-design/app/repository"
	p7service "tugas1-go/pertemuan-7-advanced-api-design/app/service"
	p7config "tugas1-go/pertemuan-7-advanced-api-design/config"
	p7database "tugas1-go/pertemuan-7-advanced-api-design/database"
	p7helper "tugas1-go/pertemuan-7-advanced-api-design/helper"
	p7route "tugas1-go/pertemuan-7-advanced-api-design/route"
)

type portfolioResponse struct {
	Data model.Portfolio `json:"data"`
}

type projectResponse struct {
	Data model.Project `json:"data"`
}

type technologyResponse struct {
	Data []model.TechnologyCount `json:"data"`
}

type storedPortfolio struct {
	StudentID int `bson:"student_id"`
	Projects  []struct {
		Technologies []string `bson:"technologies"`
	} `bson:"projects"`
}

func TestPortfolioHTTPAgainstIsolatedDatabases(t *testing.T) {
	if os.Getenv("DB_NAME") != p10config.PostgreSQLTestDatabase || os.Getenv("MONGO_DATABASE") != p10config.MongoTestDatabase {
		t.Fatalf("integration test requires DB_NAME=%s and MONGO_DATABASE=%s", p10config.PostgreSQLTestDatabase, p10config.MongoTestDatabase)
	}
	if err := p10config.ValidateTestDatabaseNames(os.Getenv("DB_NAME"), os.Getenv("MONGO_DATABASE")); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("DB_HOST") != "127.0.0.1" && os.Getenv("DB_HOST") != "localhost" {
		t.Fatal("DB_HOST must target local PostgreSQL")
	}
	if os.Getenv("DB_PORT") != "5432" {
		t.Fatal("DB_PORT must be 5432")
	}
	mongoURI := os.Getenv("MONGO_URI")
	parsedURI, err := url.Parse(mongoURI)
	if err != nil || parsedURI.Scheme != "mongodb" || (parsedURI.Hostname() != "localhost" && parsedURI.Hostname() != "127.0.0.1") || parsedURI.Port() != "27017" || (parsedURI.Path != "" && parsedURI.Path != "/portfolio_test") {
		t.Fatal("MONGO_URI must target local MongoDB port 27017 and portfolio_test")
	}
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		t.Fatal("JWT_SECRET must be configured locally with at least 32 characters")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	pool, err := p7database.NewPool(ctx)
	if err != nil {
		t.Fatalf("connect to PostgreSQL test database: %v", err)
	}
	defer pool.Close()
	var activeDatabase string
	if err := pool.QueryRow(ctx, "SELECT current_database()").Scan(&activeDatabase); err != nil || activeDatabase != p10config.PostgreSQLTestDatabase {
		t.Fatalf("active PostgreSQL database=%q, error=%v", activeDatabase, err)
	}

	users := p7repository.NewUserRepository(pool)
	owner := ensureSyntheticUser(t, ctx, users, "p10_http_owner", "p10-http-owner@example.test")
	nonOwner := ensureSyntheticUser(t, ctx, users, "p10_http_nonowner", "p10-http-nonowner@example.test")
	if owner.ID == nonOwner.ID {
		t.Fatal("owner and non-owner fixtures must be distinct")
	}
	studentID := ensureSyntheticStudent(t, ctx, pool, owner.ID)
	studentRepo := p7repository.NewStudentRepository(pool)
	student, err := studentRepo.FindByID(ctx, 1)
	if err != nil {
		t.Fatalf("read baseline synthetic student through Pertemuan 7 repository: %v", err)
	}
	if student.OwnerID != nil && *student.OwnerID != owner.ID {
		t.Fatal("student 1 already belongs to another account; refusing to change its owner")
	}
	if student.OwnerID == nil {
		command, err := pool.Exec(ctx, "UPDATE students SET owner_id=$1 WHERE id=1 AND owner_id IS NULL", owner.ID)
		if err != nil || command.RowsAffected() != 1 {
			t.Fatalf("assign synthetic owner to test student 1: rows=%d, error=%v", command.RowsAffected(), err)
		}
	}
	student, err = studentRepo.FindByID(ctx, 1)
	if err != nil || student.OwnerID == nil || *student.OwnerID != owner.ID {
		t.Fatalf("owner fixture was not visible through the application repository: student=%+v, error=%v", student, err)
	}

	client, err := p10database.NewMongoClient(ctx, mongoURI)
	if err != nil {
		t.Fatalf("connect to MongoDB test database: %v", err)
	}
	defer client.Disconnect(context.Background())
	db := client.Database(p10config.MongoTestDatabase)
	if err := p10repository.NewPortfolioRepository(db).EnsureIndexes(ctx); err != nil {
		t.Fatalf("verify portfolio unique index: %v", err)
	}
	portfolios := p10repository.NewPortfolioRepository(db)
	jwt := p7helper.NewJWTManager(secret, p7config.GetEnv("JWT_ISSUER", "praktikum-backend"), 15*time.Minute)
	roles, err := p7repository.NewRoleRepository(pool).LoadPermissions(ctx)
	if err != nil {
		t.Fatalf("load Pertemuan 7 permissions: %v", err)
	}
	permissions := p7helper.NewPermissionSet(roles)
	portfolioService := p10service.NewPortfolioService(portfolios, studentRepo, permissions)
	studentService := p7service.NewStudentService(studentRepo, permissions)
	achievementService := p7service.NewAchievementService(p7repository.NewAchievementRepository(pool))
	userService := p7service.NewUserService(users, permissions)
	authService := p7service.NewAuthService(users, p7repository.NewTokenRepository(pool), jwt, permissions, 7*24*time.Hour)
	app := p7config.NewApp(slog.New(slog.NewTextHandler(io.Discard, nil)), p7route.Dependencies{
		Pool: pool, JWT: jwt, Permissions: permissions,
		StudentService: studentService, AchievementService: achievementService,
		AuthService: authService, UserService: userService,
	}, func(app *fiber.App) {
		p10route.RegisterPortfolioRoutes(app.Group("/api/v1"), portfolioService, jwt)
	})
	defer app.Shutdown()

	ownerToken, err := jwt.GenerateAccessToken(owner)
	if err != nil {
		t.Fatalf("sign owner test token: %v", err)
	}
	nonOwnerToken, err := jwt.GenerateAccessToken(nonOwner)
	if err != nil {
		t.Fatalf("sign non-owner test token: %v", err)
	}
	basePath := "/api/v1/students/" + strconv.Itoa(studentID) + "/portfolio"

	assertHTTPStatus(t, app, http.MethodGet, basePath, "", http.StatusUnauthorized)
	assertHTTPStatus(t, app, http.MethodGet, basePath, "invalid.jwt.token", http.StatusUnauthorized)
	assertHTTPStatus(t, app, http.MethodGet, basePath, nonOwnerToken, http.StatusForbidden)
	assertHTTPStatus(t, app, http.MethodGet, "/api/v1/students/2147483647/portfolio", ownerToken, http.StatusNotFound)
	assertHTTPStatus(t, app, http.MethodGet, basePath, ownerToken, http.StatusNotFound)

	initialBody := `{"headline":"P10 integration portfolio","bio":"Synthetic HTTP integration fixture","projects":[{"title":"Initial project","technologies":["P10-Initial-Tech"],"link":"https://example.test/initial","year":2026}],"certificates":[]}`
	response := requestHTTP(t, app, http.MethodPut, basePath, ownerToken, initialBody)
	assertStatus(t, response, http.StatusCreated)
	created := decodePortfolio(t, response)
	if created.StudentID != studentID || len(created.Projects) != 1 {
		t.Fatalf("created portfolio does not match request: %+v", created)
	}
	assertPortfolioStored(t, ctx, portfolios, studentID, "P10 integration portfolio", 1)
	response = requestHTTP(t, app, http.MethodPut, basePath, nonOwnerToken, initialBody)
	assertStatus(t, response, http.StatusForbidden)
	assertPortfolioUnchanged(t, ctx, portfolios, studentID, created)

	response = requestHTTP(t, app, http.MethodGet, basePath, ownerToken, "")
	assertStatus(t, response, http.StatusOK)
	if fetched := decodePortfolio(t, response); fetched.ID != created.ID || fetched.Projects[0].ID != created.Projects[0].ID {
		t.Fatalf("GET returned a different portfolio than MongoDB stored: %+v", fetched)
	}

	replacementBody := `{"headline":"P10 replacement","bio":"Whole-resource replacement","projects":[{"title":"Replacement project","technologies":["P10-Replacement-Tech"],"link":"https://example.test/replacement","year":2026}],"certificates":[]}`
	response = requestHTTP(t, app, http.MethodPut, basePath, ownerToken, replacementBody)
	assertStatus(t, response, http.StatusOK)
	replaced := decodePortfolio(t, response)
	if replaced.Headline != "P10 replacement" || len(replaced.Projects) != 1 || replaced.Projects[0].ID == created.Projects[0].ID {
		t.Fatalf("PUT did not replace the entire representation and issue a fresh embedded ID: %+v", replaced)
	}
	assertPortfolioStored(t, ctx, portfolios, studentID, "P10 replacement", 1)

	beforeInvalid, err := portfolios.FindByStudentID(ctx, studentID)
	if err != nil {
		t.Fatal(err)
	}
	response = requestHTTP(t, app, http.MethodPut, basePath, ownerToken, `{}`)
	assertStatus(t, response, http.StatusBadRequest)
	assertPortfolioUnchanged(t, ctx, portfolios, studentID, beforeInvalid)
	response = requestHTTP(t, app, http.MethodPut, basePath, ownerToken, `{"$set":{"headline":"injected"}}`)
	if response.StatusCode < 400 || response.StatusCode >= 500 {
		t.Fatalf("NoSQL operator request status=%d, want client error", response.StatusCode)
	}
	assertPortfolioUnchanged(t, ctx, portfolios, studentID, beforeInvalid)
	response = requestHTTP(t, app, http.MethodPost, basePath+"/projects", ownerToken, `{"title":`)
	if response.StatusCode < 400 || response.StatusCode >= 500 {
		t.Fatalf("malformed JSON status=%d, want client error", response.StatusCode)
	}
	assertPortfolioUnchanged(t, ctx, portfolios, studentID, beforeInvalid)

	response = requestHTTP(t, app, http.MethodPost, basePath+"/projects", ownerToken, `{"title":"Appended project","technologies":["P10-Atomic-Phase3"],"link":"https://example.test/appended","year":2026}`)
	assertStatus(t, response, http.StatusCreated)
	added := decodeProject(t, response)
	if added.ID == "" {
		t.Fatal("POST did not return a project ID")
	}
	afterAdd, err := portfolios.FindByStudentID(ctx, studentID)
	if err != nil || len(afterAdd.Projects) != 2 {
		t.Fatalf("MongoDB state after POST has %d projects, error=%v", len(afterAdd.Projects), err)
	}
	response = requestHTTP(t, app, http.MethodDelete, basePath+"/projects/000000000000000000000000", ownerToken, "")
	assertStatus(t, response, http.StatusNotFound)
	assertPortfolioUnchanged(t, ctx, portfolios, studentID, afterAdd)

	response = requestHTTP(t, app, http.MethodGet, "/api/v1/portfolios/stats/technologies", ownerToken, "")
	assertStatus(t, response, http.StatusOK)
	var stats technologyResponse
	if err := json.NewDecoder(response.Body).Decode(&stats); err != nil {
		t.Fatalf("decode technology aggregation response: %v", err)
	}
	_ = response.Body.Close()
	expectedStats := expectedTechnologyStats(t, ctx, db.Collection("student_portfolios"))
	if fmt.Sprint(stats.Data) != fmt.Sprint(expectedStats) {
		t.Fatalf("aggregation=%v, expected from stored documents=%v", stats.Data, expectedStats)
	}

	response = requestHTTP(t, app, http.MethodDelete, basePath+"/projects/"+added.ID, ownerToken, "")
	assertStatus(t, response, http.StatusNoContent)
	afterDelete, err := portfolios.FindByStudentID(ctx, studentID)
	if err != nil || len(afterDelete.Projects) != 1 || afterDelete.Projects[0].ID == added.ID || afterDelete.Projects[0].Title != "Replacement project" {
		t.Fatalf("DELETE changed an unintended project: portfolio=%+v, error=%v", afterDelete, err)
	}

	duplicate := bson.M{"_id": bson.NewObjectID(), "student_id": studentID, "headline": "duplicate probe", "bio": "synthetic", "projects": bson.A{}, "certificates": bson.A{}}
	if _, err := db.Collection("student_portfolios").InsertOne(ctx, duplicate); err == nil || !mongo.IsDuplicateKeyError(err) {
		t.Fatalf("duplicate portfolio insert error=%v, want unique-index rejection", err)
	}
	assertPortfolioStored(t, ctx, portfolios, studentID, "P10 replacement", 1)

	projectRequests := make([]map[string]any, 20)
	for i := range projectRequests {
		projectRequests[i] = map[string]any{
			"title":        "Capacity project " + strconv.Itoa(i+1),
			"technologies": []string{"P10-Capacity-" + strconv.Itoa(i+1)},
			"link":         "https://example.test/capacity/" + strconv.Itoa(i+1),
			"year":         2026,
		}
	}
	capacityBody, err := json.Marshal(map[string]any{"headline": "P10 capacity", "bio": "Synthetic capacity test", "projects": projectRequests, "certificates": []any{}})
	if err != nil {
		t.Fatal(err)
	}
	response = requestHTTP(t, app, http.MethodPut, basePath, ownerToken, string(capacityBody))
	assertStatus(t, response, http.StatusOK)
	assertPortfolioStored(t, ctx, portfolios, studentID, "P10 capacity", 20)
	response = requestHTTP(t, app, http.MethodPost, basePath+"/projects", ownerToken, `{"title":"Overflow","technologies":["P10-Overflow"],"link":"https://example.test/overflow","year":2026}`)
	assertStatus(t, response, http.StatusConflict)
	assertPortfolioStored(t, ctx, portfolios, studentID, "P10 capacity", 20)
}

func ensureSyntheticUser(t *testing.T, ctx context.Context, users p7repository.UserRepository, username, email string) p7model.User {
	t.Helper()
	user, err := users.FindByUsername(ctx, username)
	if err == nil {
		if user.Role != "user" || !user.IsActive || user.Email != email {
			t.Fatalf("existing synthetic account %q has unexpected attributes; refusing to modify it", username)
		}
		return user
	}
	if !errors.Is(err, p7repository.ErrNotFound) {
		t.Fatalf("read synthetic account %q: %v", username, err)
	}
	password, err := p7helper.RandomToken(32)
	if err != nil {
		t.Fatalf("generate synthetic password: %v", err)
	}
	hash, err := p7helper.HashPassword(password)
	if err != nil {
		t.Fatalf("prepare synthetic password hash: %v", err)
	}
	user, err = users.Create(ctx, p7model.User{Username: username, Email: email, PasswordHash: hash, Role: "user", IsActive: true})
	if err != nil {
		t.Fatalf("create synthetic user %q: %v", username, err)
	}
	return user
}

func ensureSyntheticStudent(t *testing.T, ctx context.Context, pool *pgxpool.Pool, ownerID int) int {
	t.Helper()
	var studentID int
	var nim int64
	err := pool.QueryRow(ctx, "SELECT COALESCE(MAX(nim), 10434241123) + 1 FROM students").Scan(&nim)
	if err != nil {
		t.Fatalf("choose a fresh synthetic NIM in the test database: %v", err)
	}
	err = pool.QueryRow(ctx, `INSERT INTO students (nim,name,grade,is_active,owner_id) VALUES ($1,$2,$3,TRUE,$4) RETURNING id`, nim, "P10 HTTP Integration Student", 90.0, ownerID).Scan(&studentID)
	if err != nil {
		t.Fatalf("create synthetic HTTP integration student: %v", err)
	}
	return studentID
}

func requestHTTP(t *testing.T, app *fiber.App, method, path, token, body string) *http.Response {
	t.Helper()
	request := httptest.NewRequest(method, path, stringsReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	response, err := app.Test(request, 10000)
	if err != nil {
		t.Fatalf("HTTP %s %s: %v", method, path, err)
	}
	return response
}

func assertHTTPStatus(t *testing.T, app *fiber.App, method, path, token string, want int) {
	t.Helper()
	response := requestHTTP(t, app, method, path, token, "")
	assertStatus(t, response, want)
}

func assertStatus(t *testing.T, response *http.Response, want int) {
	t.Helper()
	if response.StatusCode != want {
		t.Fatalf("HTTP status=%d, want=%d", response.StatusCode, want)
	}
}

func decodePortfolio(t *testing.T, response *http.Response) model.Portfolio {
	t.Helper()
	defer response.Body.Close()
	var envelope portfolioResponse
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode portfolio response: %v", err)
	}
	return envelope.Data
}

func decodeProject(t *testing.T, response *http.Response) model.Project {
	t.Helper()
	defer response.Body.Close()
	var envelope projectResponse
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode project response: %v", err)
	}
	return envelope.Data
}

func assertPortfolioStored(t *testing.T, ctx context.Context, repo interface {
	FindByStudentID(context.Context, int) (model.Portfolio, error)
}, studentID int, headline string, projectCount int) {
	t.Helper()
	portfolio, err := repo.FindByStudentID(ctx, studentID)
	if err != nil {
		t.Fatalf("read portfolio after HTTP operation: %v", err)
	}
	if portfolio.Headline != headline || len(portfolio.Projects) != projectCount {
		t.Fatalf("MongoDB portfolio after HTTP operation = headline %q, %d projects; want %q, %d", portfolio.Headline, len(portfolio.Projects), headline, projectCount)
	}
}

func assertPortfolioUnchanged(t *testing.T, ctx context.Context, repo interface {
	FindByStudentID(context.Context, int) (model.Portfolio, error)
}, studentID int, before model.Portfolio) {
	t.Helper()
	after, err := repo.FindByStudentID(ctx, studentID)
	if err != nil {
		t.Fatalf("read portfolio after rejected request: %v", err)
	}
	if !reflect.DeepEqual(after, before) {
		t.Fatalf("rejected request changed stored portfolio: before=%+v after=%+v", before, after)
	}
}

func expectedTechnologyStats(t *testing.T, ctx context.Context, collection *mongo.Collection) []model.TechnologyCount {
	t.Helper()
	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		t.Fatalf("read stored portfolios to independently verify aggregation: %v", err)
	}
	defer cursor.Close(ctx)
	counts := make(map[string]int)
	for cursor.Next(ctx) {
		var portfolio storedPortfolio
		if err := cursor.Decode(&portfolio); err != nil {
			t.Fatalf("decode stored portfolio for aggregation comparison: %v", err)
		}
		for _, project := range portfolio.Projects {
			for _, technology := range project.Technologies {
				counts[technology]++
			}
		}
	}
	if err := cursor.Err(); err != nil {
		t.Fatalf("iterate stored portfolios: %v", err)
	}
	want := make([]model.TechnologyCount, 0, len(counts))
	for technology, total := range counts {
		want = append(want, model.TechnologyCount{Technology: technology, Total: total})
	}
	sort.Slice(want, func(i, j int) bool {
		if want[i].Total == want[j].Total {
			return want[i].Technology < want[j].Technology
		}
		return want[i].Total > want[j].Total
	})
	if len(want) > 10 {
		want = want[:10]
	}
	return want
}

func stringsReader(value string) *strings.Reader { return strings.NewReader(value) }
