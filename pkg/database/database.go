package database

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectToMongoDB() (*mongo.Database, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(MongoEndpoint))
	if err != nil {
		return nil, fmt.Errorf("Failed to create MongoDB client: %v", err)
	}

	err = mongoClient.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to MongoDB: %v", err)
	}

	db := mongoClient.Database(MongoDatabase)
	return db, nil
}
