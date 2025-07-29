package services

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/youtubebot/src/adapters/db"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

var jwtSecret []byte
var secretOnce sync.Once

func getJWTSecret() ([]byte, error) {
	var err error
	secretOnce.Do(func() {
		secret := os.Getenv("TOKEN")
		if len(secret) < 20 {
			err = errors.New("TOKEN is missing or too short (must be ≥32 characters)")
			return
		}
		jwtSecret = []byte(secret)
	})
	return jwtSecret, err
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req UserSignIn
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, "Invalid login request", http.StatusBadRequest)
		return
	}

	collection := db.MongoDB.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ✅ Check if email already exists
	var existing UserData
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existing)
	if err != nil {
		WriteError(w, "Invalid login credentials", http.StatusBadRequest)
		return
	}

	// Fetch user from DB and validate password...
	err = bcrypt.CompareHashAndPassword([]byte(existing.Password), []byte(req.Password))
	if err != nil {
		WriteError(w, "Invalid login credentials", http.StatusBadRequest)
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

	key, err := getJWTSecret()
	if err != nil {
		WriteError(w, "Server misconfiguration: JWT secret invalid", http.StatusInternalServerError)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(key)
	if err != nil {
		WriteError(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token": tokenString,
		"user": map[string]string{
			"id": existing.ID.Hex(),
		},
	})
}
