package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/youtubebot/src"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/download", src.HandleDownload)
	r.Get("/status/{jobID}", src.GetDownloadStatus)

	fmt.Println("🚀 Video downloader API running on :9096")
	log.Fatal(http.ListenAndServe(":9096", r))
}
