package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Connect initialises a MongoDB client with sensible defaults for connection pooling.
func Connect(ctx context.Context, uri string) (*mongo.Client, error) {
	clientOpts := options.Client().ApplyURI(uri)
	clientOpts.SetServerSelectionTimeout(5 * time.Second)
	clientOpts.SetMaxPoolSize(200)
	clientOpts.SetMinPoolSize(10)

	return mongo.Connect(ctx, clientOpts)
}
