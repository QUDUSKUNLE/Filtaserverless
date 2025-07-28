package services

import (
	"encoding/json"
	"net/http"
)

// Implement Google signin
func GoogleSignin(w http.ResponseWriter, r *http.Request) {
	// Implement Google sign-in logic

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Google sign-in is not implemented yet"})
}
