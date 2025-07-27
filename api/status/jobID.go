package handler

import (
	"net/http"
	"strings"

	"github.com/joho/godotenv"
	"github.com/youtubebot/src/adapters/db"
	"github.com/youtubebot/src/core/services"
)

func init() {
	_ = godotenv.Load()
	db.Connect()
}

func Handle(w http.ResponseWriter, r *http.Request) {
	// Extract jobID from URL manually (Vercel puts it as part of the path)
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 3 {
		http.Error(w, "Job ID not provided", http.StatusBadRequest)
		return
	}
	r = r.Clone(r.Context())
	r.URL.Path = "/" + parts[2]
	services.GetDownloadStatus(w, r)
}
