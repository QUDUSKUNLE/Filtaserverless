package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/youtubebot/src/adapters/db"
	"github.com/youtubebot/src/adapters/db/models"
	"gopkg.in/mgo.v2/bson"

	"github.com/go-chi/chi/v5"
)

func HandleDownload(w http.ResponseWriter, r *http.Request) {
	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, "❌ Invalid request body. Expecting JSON with 'url'", http.StatusBadRequest)
		return
	}
	// Generate a simple job ID
	jobID := fmt.Sprintf("job-%d", time.Now().UnixNano())
	// Background goroutine to process download
	file, err := processDownloadVideo(jobID, req)
	if err != nil {
		http.Error(w, "❌ Failed to process video: "+err.Error(), http.StatusInternalServerError)
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
		"duration":    formatDuration(file.Duration),
		"message":     "Video link is ready",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)

}

func processDownloadVideo(jobID string, req DownloadRequest) (*VideoMetadata, error) {
	log.Printf("⬇️ Starting download for job %s\n", jobID)

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
		"duration":    formatDuration(file.Duration),
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

func Home(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"Welcome": "File Downloader"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func GetDownloadStatus(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobID")
	mu.Lock()
	result, exists := downloadResults[jobID]
	mu.Unlock()

	if !exists {
		http.Error(w, "Job not found or still processing", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func GetDownloadState(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobID")
	collection := db.MongoDB.Collection("jobs")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var job models.DownloadJob
	err := collection.FindOne(ctx, bson.M{"job_id": jobID}).Decode(&job)
	if err != nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	// return JSON response
	json.NewEncoder(w).Encode(map[string]interface{}{
		"job_id": job.JobID,
		"status": job.Status,
		"url":    job.URL,
	})
}

func SignIn(w http.ResponseWriter, r *http.Request) {
	var req UserSignIn
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body. Expecting JSON with 'username' and 'password'", http.StatusBadRequest)
		return
	}

	// Simulate user sign-in logic
	user := UserResponse{ID: "user123", Username: req.Username, Email: req.Username}
	json.NewEncoder(w).Encode(user)
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	var req UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body. Expecting JSON with 'username', 'password', 'confirmPassword', 'email', 'firstName', and 'lastName'", http.StatusBadRequest)
		return
	}

	// Simulate user sign-up logic
	user := UserResponse{ID: "user123", Username: req.Username, Email: req.Email}
	json.NewEncoder(w).Encode(user)
}

// Implement Google signin
func GoogleSignin(w http.ResponseWriter, r *http.Request) {
	// Implement Google sign-in logic
	json.NewEncoder(w).Encode(map[string]string{"message": "Google sign-in is not implemented yet"})
}

// Implement Subscription
func Subscribe(w http.ResponseWriter, r *http.Request) {
	// Implement subscription logic
	json.NewEncoder(w).Encode(map[string]string{"message": "Subscription is not implemented yet"})
}
