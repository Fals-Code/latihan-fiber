package repository

import (
	"context"
	"errors"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"tugas1-go/pertemuan-10-api-mongodb/app/model"
)

func TestProjectCapacityFilterChecksPortfolioAndTwentiethSlot(t *testing.T) {
	filter := projectCapacityFilter(42)
	if filter["student_id"] != 42 {
		t.Fatalf("student_id filter = %v, want 42", filter["student_id"])
	}

	capacity, ok := filter["projects.19"].(bson.M)
	if !ok || capacity["$exists"] != false {
		t.Fatalf("capacity filter = %v, want projects.19 $exists false", filter["projects.19"])
	}
}

func TestValidateProjectCountRejectsTwentyOneProjects(t *testing.T) {
	projects := make([]model.ProjectRequest, 21)
	if err := validateProjectCount(projects); err != ErrProjectLimit {
		t.Fatalf("validateProjectCount() error = %v, want ErrProjectLimit", err)
	}
}

func TestValidateProjectCountAcceptsTwentyProjects(t *testing.T) {
	projects := make([]model.ProjectRequest, 20)
	if err := validateProjectCount(projects); err != nil {
		t.Fatalf("validateProjectCount() error = %v, want nil", err)
	}
}

func TestProjectDocumentMapsObjectIDToString(t *testing.T) {
	id := bson.NewObjectID()
	document := projectDocument{ID: id, Title: "Portfolio API"}
	project := document.toModel()
	if project.ID != id.Hex() {
		t.Fatalf("project ID = %q, want %q", project.ID, id.Hex())
	}
}

func TestClassifyMongoErrorOnlyMarksAvailabilityFailures(t *testing.T) {
	serverError := classifyMongoError("aggregate", errors.New("invalid aggregation stage"))
	if errors.Is(serverError, ErrUnavailable) {
		t.Fatal("non-network MongoDB error must not be mapped to service unavailable")
	}

	timeoutError := classifyMongoError("find", context.DeadlineExceeded)
	if !errors.Is(timeoutError, ErrUnavailable) {
		t.Fatal("operation timeout must be mapped to service unavailable")
	}
}
