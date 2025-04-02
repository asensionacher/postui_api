package database

import (
	"context"
	"fmt"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SetupMongoDB() *mongo.Collection {
	mongo_hostname := os.Getenv("MONGO_HOST")
	mongo_port := os.Getenv("MONGO_PORT")
	mongoURL := fmt.Sprintf("mongodb://%s:%s", mongo_hostname, mongo_port)
	clientOptions := options.Client().ApplyURI(mongoURL)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		panic(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		panic(err)
	}

	return client.Database("logging").Collection("logs")
}
