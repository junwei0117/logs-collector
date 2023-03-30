package database

import (
	"context"
	"fmt"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/junwei0117/logs-collector/pkg/configs"
)

var once sync.Once
var db *mongo.Database

func ConnectToMongoDB() error {
	var err error
	once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		mongoClient, err := mongo.NewClient(options.Client().ApplyURI(configs.MongoEndpoint))
		if err != nil {
			err = fmt.Errorf("Failed to create MongoDB client: %v", err)
			return
		}

		err = mongoClient.Connect(ctx)
		if err != nil {
			err = fmt.Errorf("Failed to connect to MongoDB: %v", err)
			return
		}

		db = mongoClient.Database(configs.MongoDatabase)
	})

	return err
}

func GetDB() (*mongo.Database, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}
	return db, nil
}
