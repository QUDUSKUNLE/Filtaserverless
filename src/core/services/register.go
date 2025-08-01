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
		WriteError(w, "Invalid request body. Expecting JSON with 'password', 'confirm_password', 'email', 'first_name', and 'last_name'", http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Password == "" || req.ConfirmPassword == "" || req.Email == "" || req.FirstName == "" || req.LastName == "" {
		WriteError(w, "All fields are required.", http.StatusBadRequest)
		return
	}
	if req.Password != req.ConfirmPassword {
		WriteError(w, "Passwords do not match", http.StatusBadRequest)
		return
	}
	if !strings.Contains(req.Email, "@") {
		WriteError(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	collection := db.MongoDB.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ✅ Check if email already exists
	var existing bson.M
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existing)
	if err == nil {
		WriteError(w, "User already registered", http.StatusConflict)
		return
	} else if err != mongo.ErrNoDocuments {
		log.Printf("❌ Error checking existing email: %v\n", err)
		WriteError(w, "Server error", http.StatusInternalServerError)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		WriteError(w, "Failed to hash password", http.StatusInternalServerError)
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
		WriteError(w, "Failed to get inserted ID", http.StatusInternalServerError)
		return
	}

	// Simulate user creation
	user := UserResponse{
		ID:      oid.Hex(),
		Message: "User sign up successfully",
	}
	writeJSON(w, http.StatusCreated, user)
}
