package config

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var db *mongo.Database
var RedisDB *redis.Client
var Ctx = context.Background()

func GetDBCollection(collection_name string) *mongo.Collection {
	return db.Collection(collection_name)
}

func InitDB() error {
	uri := os.Getenv("MONGO_URI")

	if uri == "" {
		return errors.New("you must set you mongodb_uri")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
		return errors.New("error while creating mongo connection")
	}

	err = client.Ping(ctx, nil); 
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	log.Println("Mongo connection successfully.")

	db = client.Database("mongo_crud")

	return nil
}

func CloseDB() error {
	return db.Client().Disconnect(context.Background())
}

func ConnectRedis() error {
	RedisDB = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	_, err := RedisDB.Ping(context.Background()).Result()

	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
		return errors.New("error while making redis connection.")
	}

	log.Println("Connected to Redis!")

	return nil
}



// config.RedisDB.Set(context.Background(), cacheKey, string(bookBytes), 10*time.Minute)

