package config

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var db *mongo.Database
var ctx context.Context
var cancel context.CancelFunc
var client *mongo.Client

func disconnect() {
	if client == nil {
		return
	}
	if err := client.Disconnect(ctx); err != nil {
		log.Fatalf("Failed to disconnect from MongoDB: %v", err)
	}
	cancel()
	log.Println("Disconnected from MongoDB")
}

func ConnectMongoDB() func() {
	mongodbURI := os.Getenv("MONGODB_URI")
	dbName := os.Getenv("MONGO_DATABASE")

	if mongodbURI == "" || dbName == "" {
		log.Fatal("MONGODB_URI or MONGO_DATABASE is not set")
	}

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)

	opts := options.Client().ApplyURI(mongodbURI).SetServerSelectionTimeout(5 * time.Second)
	c, err := mongo.Connect(ctx, opts)
	client = c
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Test the connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Ping failed: %v", err)
	}

	db = client.Database(dbName)

	log.Println("Connected to MongoDB!")
	return disconnect
}

func GetDB() *mongo.Database {
	if db == nil {
		panic("DB is not initialized")
	}
	return db
}

func GetDBCtx() (context.Context, context.CancelFunc) {
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	return ctx, cancel
}

func GetCollection(c string) *mongo.Collection {
	if db == nil {
		panic("DB is not initialized")
	}
	collection := db.Collection(c)
	return collection
}
