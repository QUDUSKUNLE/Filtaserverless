package handler

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/youtubebot/src/adapters/db"
	middle "github.com/youtubebot/src/adapters/middleware"
	"github.com/youtubebot/src/core/services"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	_ = godotenv.Load()
	db.Connect()
	EnsureUserIndexes()
}

func EnsureUserIndexes() {
	collection := db.MongoDB.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a unique index on the email field
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}}, // index on email field
		Options: options.Index().SetUnique(true),
	}

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Printf("⚠️ Failed to create unique index on email: %v", err)
	} else {
		log.Println("✅ Unique index on email ensured")
	}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	middle.CorsMiddleware(http.HandlerFunc(services.SignUp)).ServeHTTP(w, r)
}
