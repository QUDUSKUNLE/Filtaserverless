package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/youtubebot/src/adapters/db"
	"go.mongodb.org/mongo-driver/bson"
)

func HandleDownload(w http.ResponseWriter, r *http.Request) {
	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		writeError(w, "❌ Invalid request body. Expecting JSON with 'url'", http.StatusBadRequest)
		return
	}
	// Generate a simple job ID
	jobID := fmt.Sprintf("job-%d", time.Now().UnixNano())
	// Background goroutine to process download
	file, err := processDownloadVideo(jobID, req)
	if err != nil {
		writeError(w, "❌ Failed to process video: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Immediately respond that job is accepted
	resp := map[string]string{
		"job_id":      jobID,
		"direct_link": file.URL,
		"title":       file.Title,
		"description": file.Description,
		"thumbnail":   file.Thumbnail,
		"webpage_url": file.WebpageURL,
		"extension":   file.Ext,
		"format_id":   file.FormatID,
		"filesize":    formatSize(file.Filesize),
		"duration":    formatDuration(int64(file.Duration)),
		"message":     "Video link is ready",
	}
	writeJSON(w, http.StatusOK, resp)
}

func processDownloadVideo(jobID string, req DownloadRequest) (*VideoMetadata, error) {
	log.Printf("⬇️ Starting fetch for job %s\n", jobID)

	file, err := getDirectDownloadURL(req.URL)
	if err != nil {
		log.Printf("❌ Job %s failed: %v\n", jobID, err)
		return nil, err
	}

	job := bson.M{
		"job_id":      jobID,
		"url":         req.URL,
		"directory":   file.URL,
		"status":      "success",
		"direct_link": file.URL,
		"title":       file.Title,
		"description": file.Description,
		"thumbnail":   file.Thumbnail,
		"webpage_url": file.WebpageURL,
		"extension":   file.Ext,
		"format_id":   file.FormatID,
		"filesize":    formatSize(file.Filesize),
		"duration":    formatDuration(int64(file.Duration)),
		"created_at":  time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.MongoDB.Collection("jobs")
	if _, err := collection.InsertOne(ctx, job); err != nil {
		log.Printf("❌ Failed to insert job %s: %v\n", jobID, err)
		return nil, err
	}

	log.Printf("✅ Job %s completed. File saved to: %s\n", jobID, file.Title)
	return file, nil
}
