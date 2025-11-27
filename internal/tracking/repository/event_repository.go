package repository

import (
	"context"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/tdevilphan/quote-snap-golang/internal/tracking/domain"
)

// EventRepository defines persistence operations required by the service domain.
type EventRepository interface {
	Persist(ctx context.Context, event domain.Event) error
}

// MongoEventRepository stores events inside MongoDB with bounded indexes.
type MongoEventRepository struct {
	collection *mongo.Collection
}

// NewMongoEventRepository wires a Mongo collection into a repository implementation.
func NewMongoEventRepository(db *mongo.Database) (*MongoEventRepository, error) {
	collection := db.Collection("events")
	if err := ensureIndexes(context.Background(), collection); err != nil {
		return nil, errors.Wrap(err, "ensure indexes")
	}
	return &MongoEventRepository{collection: collection}, nil
}

// Persist writes a single event document.
func (r *MongoEventRepository) Persist(ctx context.Context, event domain.Event) error {
	_, err := r.collection.InsertOne(ctx, bson.M{
		"_id":         event.ID.String(),
		"name":        event.Name,
		"user_id":     event.UserID,
		"source":      event.Source,
		"metadata":    event.Metadata,
		"occurred_at": event.OccurredAt,
		"received_at": event.ReceivedAt,
	})
	return errors.Wrap(err, "insert event")
}

func ensureIndexes(ctx context.Context, collection *mongo.Collection) error {
	model := mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "occurred_at", Value: -1},
		},
		Options: options.Index().SetBackground(true),
	}
	_, err := collection.Indexes().CreateOne(ctx, model)
	return err
}
