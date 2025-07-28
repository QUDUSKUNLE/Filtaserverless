package handler

import (
	"net/http"

	"github.com/joho/godotenv"
	"github.com/youtubebot/src/adapters/db"
	middle "github.com/youtubebot/src/adapters/middleware"
	"github.com/youtubebot/src/core/services"
)

func init() {
	_ = godotenv.Load()
	db.Connect()
}

func Handler(w http.ResponseWriter, r *http.Request) {
	middle.CorsMiddleware(http.HandlerFunc(services.HandleDownload)).ServeHTTP(w, r)
}
