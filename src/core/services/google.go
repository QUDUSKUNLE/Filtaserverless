package services

import (
	"net/http"
)

// Implement Google signin
func GoogleSignin(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"message": "Google sign-in is not implemented yet"})
}
