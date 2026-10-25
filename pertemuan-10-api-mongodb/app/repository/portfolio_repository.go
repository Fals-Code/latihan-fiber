package repository

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"tugas1-go/pertemuan-10-api-mongodb/app/model"
)

const maxProjectsPerPortfolio = 20

var (
	ErrNotFound     = errors.New("portfolio or project not found")
	ErrDuplicate    = errors.New("portfolio already exists")
	ErrProjectLimit = errors.New("portfolio project limit reached")
	ErrInvalidID    = errors.New("project ID is invalid")
	ErrUnavailable  = errors.New("MongoDB is unavailable")
)

type PortfolioRepository interface {
	EnsureIndexes(ctx context.Context) error
	FindByStudentID(ctx context.Context, studentID int) (model.Portfolio, error)
	Replace(ctx context.Context, studentID int, request model.PortfolioRequest) (model.Portfolio, bool, error)
	AddProject(ctx context.Context, studentID int, request model.ProjectRequest) (model.Project, error)
	DeleteProject(ctx context.Context, studentID int, projectID string) error
	TopTechnologies(ctx context.Context) ([]model.TechnologyCount, error)
}

type portfolioMongoRepository struct {
	collection *mongo.Collection
}

type portfolioDocument struct {
	ID           bson.ObjectID         `bson:"_id"`
	StudentID    int                   `bson:"student_id"`
	Headline     string                `bson:"headline"`
	Bio          string                `bson:"bio"`
	Projects     []projectDocument     `bson:"projects"`
	Certificates []certificateDocument `bson:"certificates"`
}

type projectDocument struct {
	ID           bson.ObjectID `bson:"_id"`
	Title        string        `bson:"title"`
	Technologies []string      `bson:"technologies"`
	Link         string        `bson:"link"`
	Year         int           `bson:"year"`
}

type certificateDocument struct {
	Name   string    `bson:"name"`
	Issuer string    `bson:"issuer"`
	Date   time.Time `bson:"date"`
}

type technologyCountDocument struct {
	Technology string `bson:"_id"`
	Total      int    `bson:"total"`
}

func NewPortfolioRepository(db *mongo.Database) PortfolioRepository {
	return &portfolioMongoRepository{collection: db.Collection("student_portfolios")}
}

func (r *portfolioMongoRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "student_id", Value: 1}},
		Options: options.Index().
			SetUnique(true).
			SetName("student_id_unique"),
	})
	if err != nil {
		return classifyMongoError("create student portfolio index", err)
	}
	return nil
}

func (r *portfolioMongoRepository) FindByStudentID(ctx context.Context, studentID int) (model.Portfolio, error) {
	var document portfolioDocument
	err := r.collection.FindOne(ctx, bson.M{"student_id": studentID}).Decode(&document)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return model.Portfolio{}, ErrNotFound
	}
	if err != nil {
		return model.Portfolio{}, classifyMongoError("find student portfolio", err)
	}
	return document.toModel(), nil
}

func (r *portfolioMongoRepository) Replace(
	ctx context.Context,
	studentID int,
	request model.PortfolioRequest,
) (model.Portfolio, bool, error) {
	if err := validateProjectCount(request.Projects); err != nil {
		return model.Portfolio{}, false, err
	}

	projects := make([]projectDocument, 0, len(request.Projects))
	for _, project := range request.Projects {
		projects = append(projects, newProjectDocument(project))
	}
	certificates := make([]certificateDocument, 0, len(request.Certificates))
	for _, certificate := range request.Certificates {
		certificates = append(certificates, certificateDocument{
			Name: certificate.Name, Issuer: certificate.Issuer, Date: certificate.Date,
		})
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"student_id": studentID},
		bson.M{
			"$set": bson.M{
				"headline":     request.Headline,
				"bio":          request.Bio,
				"projects":     projects,
				"certificates": certificates,
			},
			"$setOnInsert": bson.M{
				"_id":        bson.NewObjectID(),
				"student_id": studentID,
			},
		},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return model.Portfolio{}, false, ErrDuplicate
		}
		return model.Portfolio{}, false, classifyMongoError("replace student portfolio", err)
	}

	portfolio, err := r.FindByStudentID(ctx, studentID)
	if err != nil {
		return model.Portfolio{}, false, err
	}
	return portfolio, result.UpsertedCount == 1, nil
}

func (r *portfolioMongoRepository) AddProject(
	ctx context.Context,
	studentID int,
	request model.ProjectRequest,
) (model.Project, error) {
	project := newProjectDocument(request)
	var updated portfolioDocument
	err := r.collection.FindOneAndUpdate(
		ctx,
		projectCapacityFilter(studentID),
		bson.M{"$push": bson.M{"projects": project}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)
	if err == nil {
		for _, saved := range updated.Projects {
			if saved.ID == project.ID {
				return saved.toModel(), nil
			}
		}
		return model.Project{}, fmt.Errorf("added project was not returned in portfolio")
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return model.Project{}, classifyMongoError("add portfolio project", err)
	}

	count, err := r.collection.CountDocuments(ctx, bson.M{"student_id": studentID})
	if err != nil {
		return model.Project{}, classifyMongoError("check student portfolio", err)
	}
	if count == 0 {
		return model.Project{}, ErrNotFound
	}
	return model.Project{}, ErrProjectLimit
}

func (r *portfolioMongoRepository) DeleteProject(
	ctx context.Context,
	studentID int,
	projectID string,
) error {
	objectID, err := bson.ObjectIDFromHex(projectID)
	if err != nil {
		return ErrInvalidID
	}

	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"student_id": studentID, "projects._id": objectID},
		bson.M{"$pull": bson.M{"projects": bson.M{"_id": objectID}}},
	)
	if err != nil {
		return classifyMongoError("delete portfolio project", err)
	}
	if result.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *portfolioMongoRepository) TopTechnologies(ctx context.Context) ([]model.TechnologyCount, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$unwind", Value: "$projects"}},
		bson.D{{Key: "$unwind", Value: "$projects.technologies"}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$projects.technologies"},
			{Key: "total", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		bson.D{{Key: "$sort", Value: bson.D{
			{Key: "total", Value: -1},
			{Key: "_id", Value: 1},
		}}},
		bson.D{{Key: "$limit", Value: 10}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, classifyMongoError("aggregate portfolio technologies", err)
	}
	defer cursor.Close(ctx)

	var documents []technologyCountDocument
	if err := cursor.All(ctx, &documents); err != nil {
		return nil, classifyMongoError("decode portfolio technology statistics", err)
	}

	result := make([]model.TechnologyCount, 0, len(documents))
	for _, document := range documents {
		result = append(result, model.TechnologyCount{
			Technology: document.Technology,
			Total:      document.Total,
		})
	}
	return result, nil
}

func classifyMongoError(operation string, err error) error {
	if mongo.IsNetworkError(err) || mongo.IsTimeout(err) || errors.Is(err, mongo.ErrClientDisconnected) {
		return fmt.Errorf("%w: %s: %v", ErrUnavailable, operation, err)
	}
	var networkError net.Error
	if errors.As(err, &networkError) {
		return fmt.Errorf("%w: %s: %v", ErrUnavailable, operation, err)
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func projectCapacityFilter(studentID int) bson.M {
	return bson.M{
		"student_id":  studentID,
		"projects.19": bson.M{"$exists": false},
	}
}

func validateProjectCount(projects []model.ProjectRequest) error {
	if len(projects) > maxProjectsPerPortfolio {
		return ErrProjectLimit
	}
	return nil
}

func newProjectDocument(project model.ProjectRequest) projectDocument {
	technologies := append([]string{}, project.Technologies...)
	return projectDocument{
		ID:           bson.NewObjectID(),
		Title:        project.Title,
		Technologies: technologies,
		Link:         project.Link,
		Year:         project.Year,
	}
}

func (document portfolioDocument) toModel() model.Portfolio {
	projects := make([]model.Project, 0, len(document.Projects))
	for _, project := range document.Projects {
		projects = append(projects, project.toModel())
	}

	certificates := make([]model.Certificate, 0, len(document.Certificates))
	for _, certificate := range document.Certificates {
		certificates = append(certificates, model.Certificate{
			Name: certificate.Name, Issuer: certificate.Issuer, Date: certificate.Date,
		})
	}

	return model.Portfolio{
		ID:           document.ID.Hex(),
		StudentID:    document.StudentID,
		Headline:     document.Headline,
		Bio:          document.Bio,
		Projects:     projects,
		Certificates: certificates,
	}
}

func (document projectDocument) toModel() model.Project {
	technologies := append([]string{}, document.Technologies...)
	return model.Project{
		ID:           document.ID.Hex(),
		Title:        document.Title,
		Technologies: technologies,
		Link:         document.Link,
		Year:         document.Year,
	}
}
