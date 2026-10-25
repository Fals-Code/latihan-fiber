//go:build integration

package database

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"tugas1-go/pertemuan-10-api-mongodb/config"
)

func TestMongoPortfolioDatabasePreflightReadOnly(t *testing.T) {
	mongoName := os.Getenv("MONGO_DATABASE")
	if mongoName == "" || mongoName != config.MongoTestDatabase {
		t.Fatalf("MONGO_DATABASE must explicitly equal %q", config.MongoTestDatabase)
	}
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		t.Fatal("MONGO_URI must be explicitly set for integration preflight")
	}
	parsedURI, err := url.Parse(mongoURI)
	if err != nil || parsedURI.Scheme != "mongodb" || (parsedURI.Hostname() != "localhost" && parsedURI.Hostname() != "127.0.0.1") || parsedURI.Port() != "27017" {
		t.Fatal("MONGO_URI must target the approved local MongoDB server at port 27017")
	}
	if parsedURI.Path != "" && parsedURI.Path != "/portfolio_test" {
		t.Fatal("MONGO_URI cannot select a database outside portfolio_test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := NewMongoClient(ctx, mongoURI)
	if err != nil {
		if client != nil {
			_ = client.Disconnect(context.Background())
		}
		t.Fatalf("connect to local MongoDB test server: %v", err)
	}
	defer client.Disconnect(context.Background())
	db := client.Database(mongoName)
	if db.Name() != config.MongoTestDatabase {
		t.Fatalf("selected MongoDB database=%q, want %q", db.Name(), config.MongoTestDatabase)
	}
	collections, err := db.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		t.Fatalf("list collections in MongoDB test database: %v", err)
	}
	if !containsPortfolioCollection(collections, "student_portfolios") {
		t.Fatal("student_portfolios collection is missing; read-only preflight will not create it")
	}

	indexes, err := db.Collection("student_portfolios").Indexes().List(ctx)
	if err != nil {
		t.Fatalf("list portfolio indexes: %v", err)
	}
	defer indexes.Close(ctx)
	for indexes.Next(ctx) {
		var index struct {
			Name   string `bson:"name"`
			Key    bson.M `bson:"key"`
			Unique bool   `bson:"unique"`
		}
		if err := indexes.Decode(&index); err != nil {
			t.Fatalf("decode portfolio index metadata: %v", err)
		}
		if index.Name == "student_id_unique" && index.Unique && fmt.Sprint(index.Key["student_id"]) == "1" {
			return
		}
	}
	if err := indexes.Err(); err != nil {
		t.Fatalf("read portfolio index metadata: %v", err)
	}
	t.Fatal("student_id_unique unique index is missing; read-only preflight will not create it")
}

func containsPortfolioCollection(collections []string, wanted string) bool {
	for _, collection := range collections {
		if collection == wanted {
			return true
		}
	}
	return false
}
