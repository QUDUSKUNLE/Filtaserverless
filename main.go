package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/youtubebot/src/adapters/db"
	middle "github.com/youtubebot/src/adapters/middleware"
	"github.com/youtubebot/src/core/services"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	r := chi.NewRouter()
	r.Use(middle.CorsMiddleware)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	db.Connect()

	r.Get("/", services.Home)
	r.Post("/analyse", services.HandleDownload)
	r.Get("/status/{jobID}", services.GetDownloadStatus)

	fmt.Println("🚀 Filta running on :9096")
	log.Fatal(http.ListenAndServe(":9096", r))
}
