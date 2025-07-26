package db

import (
	"context"
	"fmt"
	"log"
	"os"
	// "time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client
var MongoDB *mongo.Database

func Connect() {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("MONGODB_URI not set in environment")
	}

	opts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		log.Fatalf("MongoDB connect error: %v", err)
	}

	if err := client.Ping(context.TODO(), nil); err != nil {
		log.Fatalf("MongoDB ping error: %v", err)
	}

	MongoClient = client
	MongoDB = client.Database("youtubebot")
	fmt.Println("You successfully connected to MongoDB!")
}
