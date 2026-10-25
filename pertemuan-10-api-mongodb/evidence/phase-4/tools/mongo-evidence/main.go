package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const databaseName = "portfolio_test"

type project struct {
	ID           bson.ObjectID `bson:"_id" json:"id"`
	Title        string        `bson:"title" json:"title"`
	Technologies []string      `bson:"technologies" json:"technologies"`
	Link         string        `bson:"link" json:"link"`
	Year         int           `bson:"year" json:"year"`
}

type portfolio struct {
	ID           bson.ObjectID `bson:"_id" json:"id"`
	StudentID    int           `bson:"student_id" json:"student_id"`
	Headline     string        `bson:"headline" json:"headline"`
	Bio          string        `bson:"bio" json:"bio"`
	Projects     []project     `bson:"projects" json:"projects"`
	Certificates []struct {
		Name   string    `bson:"name" json:"name"`
		Issuer string    `bson:"issuer" json:"issuer"`
		Date   time.Time `bson:"date" json:"date"`
	} `bson:"certificates" json:"certificates"`
}

type technologyCount struct {
	Technology string `json:"technology"`
	Total      int    `json:"total"`
}

func main() {
	if len(os.Args) < 2 {
		fail("usage: go run mongo-evidence.go <snapshot|portfolio|stats|explain|indexes> [studentID]")
	}
	uri := os.Getenv("MONGO_URI")
	name := os.Getenv("MONGO_DATABASE")
	if name != databaseName || uri != "mongodb://127.0.0.1:27017" {
		fail("MONGO_URI and MONGO_DATABASE must explicitly target local portfolio_test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(options.Client().ApplyURI(uri).SetServerSelectionTimeout(5 * time.Second))
	if err != nil {
		fail("connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())
	if err := client.Ping(ctx, nil); err != nil {
		fail("ping MongoDB: %v", err)
	}
	db := client.Database(databaseName)
	collection := db.Collection("student_portfolios")

	switch os.Args[1] {
	case "snapshot", "portfolio", "explain":
		studentID := parseStudentID()
		switch os.Args[1] {
		case "snapshot":
			snapshot(ctx, collection, studentID)
		case "portfolio":
			showPortfolio(ctx, collection, studentID)
		case "explain":
			explain(ctx, db, studentID)
		}
	case "stats":
		showIndependentStats(ctx, collection)
	case "indexes":
		showIndexes(ctx, collection)
	default:
		fail("unknown mode %q", os.Args[1])
	}
}

func parseStudentID() int {
	if len(os.Args) < 3 {
		fail("studentID argument is required")
	}
	id, err := strconv.Atoi(os.Args[2])
	if err != nil || id < 1 {
		fail("studentID must be a positive integer")
	}
	return id
}

func snapshot(ctx context.Context, collection *mongo.Collection, studentID int) {
	var document bson.D
	err := collection.FindOne(ctx, bson.D{{Key: "student_id", Value: studentID}}).Decode(&document)
	if err != nil {
		fail("find portfolio: %v", err)
	}
	encoded, err := bson.MarshalExtJSON(document, false, false)
	if err != nil {
		fail("encode portfolio snapshot: %v", err)
	}
	digest := sha256.Sum256(encoded)
	fmt.Printf("student_id=%d sha256=%s bytes=%d\n", studentID, hex.EncodeToString(digest[:]), len(encoded))
}

func showPortfolio(ctx context.Context, collection *mongo.Collection, studentID int) {
	var document portfolio
	err := collection.FindOne(ctx, bson.D{{Key: "student_id", Value: studentID}}).Decode(&document)
	if err != nil {
		fail("find portfolio: %v", err)
	}
	uniqueIDs := make(map[bson.ObjectID]struct{}, len(document.Projects))
	validProjects := true
	for _, item := range document.Projects {
		if item.ID.IsZero() || item.Title == "" || len(item.Technologies) == 0 || item.Link == "" || item.Year < 1 {
			validProjects = false
		}
		if _, exists := uniqueIDs[item.ID]; exists {
			validProjects = false
		}
		uniqueIDs[item.ID] = struct{}{}
	}
	output := struct {
		portfolio
		ProjectCount       int  `json:"project_count"`
		ProjectIDsUnique   bool `json:"project_ids_unique"`
		ProjectFieldsValid bool `json:"project_fields_valid"`
	}{document, len(document.Projects), len(uniqueIDs) == len(document.Projects), validProjects}
	writeJSON(output)
}

func showIndependentStats(ctx context.Context, collection *mongo.Collection) {
	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		fail("read stored portfolios: %v", err)
	}
	defer cursor.Close(ctx)
	counts := make(map[string]int)
	for cursor.Next(ctx) {
		var document portfolio
		if err := cursor.Decode(&document); err != nil {
			fail("decode stored portfolio: %v", err)
		}
		for _, item := range document.Projects {
			for _, technology := range item.Technologies {
				counts[technology]++
			}
		}
	}
	if err := cursor.Err(); err != nil {
		fail("iterate stored portfolios: %v", err)
	}
	result := make([]technologyCount, 0, len(counts))
	for technology, total := range counts {
		result = append(result, technologyCount{Technology: technology, Total: total})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Total == result[j].Total {
			return result[i].Technology < result[j].Technology
		}
		return result[i].Total > result[j].Total
	})
	if len(result) > 10 {
		result = result[:10]
	}
	writeJSON(result)
}

func explain(ctx context.Context, db *mongo.Database, studentID int) {
	command := bson.D{
		{Key: "explain", Value: bson.D{
			{Key: "find", Value: "student_portfolios"},
			{Key: "filter", Value: bson.D{{Key: "student_id", Value: studentID}}},
		}},
		{Key: "verbosity", Value: "executionStats"},
	}
	var result bson.M
	if err := db.RunCommand(ctx, command).Decode(&result); err != nil {
		fail("run executionStats explain: %v", err)
	}
	writeJSON(result)
}

func showIndexes(ctx context.Context, collection *mongo.Collection) {
	cursor, err := collection.Indexes().List(ctx)
	if err != nil {
		fail("list collection indexes: %v", err)
	}
	defer cursor.Close(ctx)
	var indexes []bson.M
	if err := cursor.All(ctx, &indexes); err != nil {
		fail("decode collection indexes: %v", err)
	}
	writeJSON(indexes)
}

func writeJSON(value any) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		fail("write JSON: %v", err)
	}
}

func fail(format string, args ...any) {
	if len(args) > 0 && errors.Is(fmt.Errorf(format, args...), mongo.ErrClientDisconnected) {
		fmt.Fprintln(os.Stderr, "MongoDB client is unavailable")
	} else {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
	os.Exit(1)
}
