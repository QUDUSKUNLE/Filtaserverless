package services

import (
	"net/http"
)

// Implement Subscription
func Subscribe(w http.ResponseWriter, r *http.Request) {
	// Implement subscription logic
	writeJSON(w, http.StatusOK, map[string]string{"message": "Subscription is not implemented yet"})
}
