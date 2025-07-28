package services

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/youtubebot/src/adapters/db"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"
)

func SignUp(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body. Expecting JSON with 'username', 'password', 'confirmPassword', 'email', 'firstName', and 'lastName'", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Password == "" || req.ConfirmPassword == "" || req.Email == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}
	if req.Password != req.ConfirmPassword {
		http.Error(w, "Passwords do not match", http.StatusBadRequest)
		return
	}
	if !strings.Contains(req.Email, "@") {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	collection := db.MongoDB.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ✅ Check if email already exists
	var existing bson.M
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existing)
	if err == nil {
		http.Error(w, "User already registered", http.StatusConflict)
		return
	} else if err != mongo.ErrNoDocuments {
		log.Printf("❌ Error checking existing email: %v\n", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	register := bson.M{
		"username":   req.Username,
		"password":   string(hashedPassword),
		"email":      req.Email,
		"first_name": req.FirstName,
		"last_name":  req.LastName,
	}

	result, err := collection.InsertOne(ctx, register)
	if err != nil {
		log.Printf("❌ Failed to insert user: %v\n", err)
		return
	}

	// Assert ObjectID and convert to hex string
	oid, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		http.Error(w, "Failed to get inserted ID", http.StatusInternalServerError)
		return
	}

	// Simulate user creation
	user := UserResponse{
		ID:      oid.Hex(),
		Message: "User sign up successfully",
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}
