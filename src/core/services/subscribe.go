package services

import (
	"encoding/json"
	"net/http"
)

// Implement Subscription
func Subscribe(w http.ResponseWriter, r *http.Request) {
	// Implement subscription logic

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Subscription is not implemented yet"})
}
