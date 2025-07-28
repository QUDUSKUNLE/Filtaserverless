package services

import (
	"net/http"
)

func Home(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"Welcome": "Filta Downloader"}
	writeJSON(w, http.StatusOK, resp)
}
