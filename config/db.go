package config

import (
	"context"
	"errors"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var db *mongo.Database

func GetDBCollection(collection_name string) *mongo.Collection {
	return db.Collection(collection_name)
}

func InitDB() error {
	uri := os.Getenv("MONGO_URI")

	if uri == "" {
		return errors.New("you must set you mongodb_uri")
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))

	if err != nil {
		return errors.New("error while creating mongo connection")
	}

	db = client.Database("mongo_crud")

	return nil
}

func CloseDB() error {
	return db.Client().Disconnect(context.Background())
}
