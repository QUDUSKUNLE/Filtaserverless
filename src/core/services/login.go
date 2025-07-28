package services

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/youtubebot/src/adapters/db"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

func Login(w http.ResponseWriter, r *http.Request) {

	jwtSecret := []byte(os.Getenv("TOKEN"))
	var req UserSignIn
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid login request", http.StatusBadRequest)
		return
	}

	collection := db.MongoDB.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ✅ Check if email already exists
	var existing UserData
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existing)
	if err != nil {
		http.Error(w, "Invalid login credentials", http.StatusBadRequest)
		return
	}

	// Fetch user from DB and validate password...
	err = bcrypt.CompareHashAndPassword([]byte(existing.Password), []byte(req.Password))
	if err != nil {
		http.Error(w, "Invalid login credentials", http.StatusBadRequest)
		return
	}

	// Generate token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Email: existing.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   existing.ID.Hex(), // Use user's ObjectID as subject
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token": tokenString,
		"user": map[string]string{
			"id": existing.ID.Hex(),
		},
	})
}
