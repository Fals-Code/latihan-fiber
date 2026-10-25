//go:build integration

package repository

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"tugas1-go/pertemuan-10-api-mongodb/app/model"
	"tugas1-go/pertemuan-10-api-mongodb/config"
	p10database "tugas1-go/pertemuan-10-api-mongodb/database"
	p7repository "tugas1-go/pertemuan-7-advanced-api-design/app/repository"
	p7database "tugas1-go/pertemuan-7-advanced-api-design/database"
)

const (
	testStudentName    = "P10 Test Student"
	testPortfolioTitle = "P10 Synthetic Portfolio"
)

func TestPortfolioRepositoryAgainstIsolatedDatabases(t *testing.T) {
	postgresName := requiredEnvironmentValue(t, "DB_NAME")
	mongoName := requiredEnvironmentValue(t, "MONGO_DATABASE")
	if err := config.ValidateTestDatabaseNames(postgresName, mongoName); err != nil {
		t.Fatal(err)
	}
	if host := requiredEnvironmentValue(t, "DB_HOST"); host != "localhost" && host != "127.0.0.1" {
		t.Fatalf("DB_HOST must point to the local PostgreSQL server, got %q", host)
	}
	if port := requiredEnvironmentValue(t, "DB_PORT"); port != "5432" {
		t.Fatalf("DB_PORT must be 5432 for the local test server, got %q", port)
	}

	mongoURI := requiredEnvironmentValue(t, "MONGO_URI")
	parsedURI, err := url.Parse(mongoURI)
	if err != nil || parsedURI.Scheme != "mongodb" || (parsedURI.Hostname() != "localhost" && parsedURI.Hostname() != "127.0.0.1") || parsedURI.Port() != "27017" {
		t.Fatal("MONGO_URI must explicitly target the local MongoDB server at port 27017")
	}
	if parsedURI.Path != "" && parsedURI.Path != "/portfolio_test" {
		t.Fatal("MONGO_URI must not select a different database")
	}

	studentID, err := strconv.Atoi(requiredEnvironmentValue(t, "P10_TEST_STUDENT_ID"))
	if err != nil || studentID <= 0 {
		t.Fatal("P10_TEST_STUDENT_ID must be a positive integer for the synthetic student fixture")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	postgresPool, err := p7database.NewPool(ctx)
	if err != nil {
		t.Fatalf("connect to isolated PostgreSQL test database: %v", err)
	}
	defer postgresPool.Close()

	var activeDatabase string
	if err := postgresPool.QueryRow(ctx, "SELECT current_database()").Scan(&activeDatabase); err != nil {
		t.Fatalf("read active PostgreSQL database: %v", err)
	}
	if activeDatabase != config.PostgreSQLTestDatabase {
		t.Fatalf("connected to PostgreSQL database %q, want %q", activeDatabase, config.PostgreSQLTestDatabase)
	}

	student, err := p7repository.NewStudentRepository(postgresPool).FindByID(ctx, studentID)
	if err != nil {
		t.Fatalf("read synthetic student through the Pertemuan 7 repository: %v", err)
	}
	if student.Name != testStudentName {
		t.Fatalf("student name = %q, want synthetic fixture %q", student.Name, testStudentName)
	}

	client, err := p10database.NewMongoClient(ctx, mongoURI)
	if err != nil {
		t.Fatalf("connect to local MongoDB test server: %v", err)
	}
	defer client.Disconnect(context.Background())

	db := client.Database(mongoName)
	if db.Name() != config.MongoTestDatabase {
		t.Fatalf("MongoDB database = %q, want %q", db.Name(), config.MongoTestDatabase)
	}
	collectionNames, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		t.Fatalf("list MongoDB test collections: %v", err)
	}
	if !containsString(collectionNames, "student_portfolios") {
		if err := db.CreateCollection(ctx, "student_portfolios"); err != nil {
			t.Fatalf("create student_portfolios test collection: %v", err)
		}
	}

	repository := NewPortfolioRepository(db)
	if err := repository.EnsureIndexes(ctx); err != nil {
		t.Fatalf("ensure unique student_id index: %v", err)
	}
	verifyStudentIDUniqueIndex(t, ctx, db)

	portfolio, err := repository.FindByStudentID(ctx, studentID)
	if errors.Is(err, ErrNotFound) {
		request := model.PortfolioRequest{
			Headline:     testPortfolioTitle,
			Bio:          "Data sintetis untuk verifikasi integrasi Pertemuan 10.",
			Projects:     []model.ProjectRequest{},
			Certificates: []model.CertificateRequest{},
		}
		if _, created, replaceErr := repository.Replace(ctx, studentID, request); replaceErr != nil {
			t.Fatalf("create synthetic MongoDB portfolio through the repository: %v", replaceErr)
		} else if !created {
			t.Fatal("synthetic portfolio was not created in the empty test database")
		}
		portfolio, err = repository.FindByStudentID(ctx, studentID)
	}
	if err != nil {
		t.Fatalf("read portfolio through the repository: %v", err)
	}
	if portfolio.Headline != testPortfolioTitle {
		t.Fatalf("existing test portfolio headline = %q; refusing to overwrite non-fixture data", portfolio.Headline)
	}

	collection := db.Collection("student_portfolios")
	duplicate := bson.M{
		"_id":          bson.NewObjectID(),
		"student_id":   studentID,
		"headline":     "duplicate must be rejected",
		"bio":          "synthetic duplicate-index probe",
		"projects":     bson.A{},
		"certificates": bson.A{},
	}
	if _, err := collection.InsertOne(ctx, duplicate); err == nil || !mongo.IsDuplicateKeyError(err) {
		t.Fatalf("second document for student_id %d error = %v, want duplicate-key rejection", studentID, err)
	}

	count, err := collection.CountDocuments(ctx, bson.M{"student_id": studentID})
	if err != nil {
		t.Fatalf("count portfolios after duplicate-key probe: %v", err)
	}
	if count != 1 {
		t.Fatalf("portfolio count for synthetic student = %d, want 1", count)
	}
}

func requiredEnvironmentValue(t *testing.T, name string) string {
	t.Helper()
	value, exists := os.LookupEnv(name)
	if !exists || value == "" {
		t.Fatalf("%s must be explicitly set before the integration test can access a database", name)
	}
	return value
}

func verifyStudentIDUniqueIndex(t *testing.T, ctx context.Context, db *mongo.Database) {
	t.Helper()
	cursor, err := db.Collection("student_portfolios").Indexes().List(ctx)
	if err != nil {
		t.Fatalf("list student_portfolios indexes: %v", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var index struct {
			Name   string `bson:"name"`
			Key    bson.M `bson:"key"`
			Unique bool   `bson:"unique"`
		}
		if err := cursor.Decode(&index); err != nil {
			t.Fatalf("decode MongoDB index: %v", err)
		}
		if index.Name == "student_id_unique" && index.Unique && fmt.Sprint(index.Key["student_id"]) == "1" {
			return
		}
	}
	if err := cursor.Err(); err != nil {
		t.Fatalf("read MongoDB indexes: %v", err)
	}
	t.Fatal("student_id_unique index with unique=true is not present")
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
