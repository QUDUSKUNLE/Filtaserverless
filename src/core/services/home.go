package services

import (
	"encoding/json"
	"net/http"
)

func Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]string{"Welcome": "Filta Downloader"}
	json.NewEncoder(w).Encode(resp)
}
