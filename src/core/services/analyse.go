package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"github.com/youtubebot/src/adapters/db/models"
	"github.com/youtubebot/src/adapters/db/repository"
)

func Analyse(w http.ResponseWriter, r *http.Request) {
	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		WriteError(w, "❌ Invalid request body. Expecting JSON with 'url'", http.StatusBadRequest)
		return
	}

	// user := GetUserID(r)
	// if user == "" {
	// 	WriteError(w, "Unauthorized to perform operation", http.StatusUnauthorized)
	// }
	// req.UserID = user
	// Generate a simple job ID
	jobID := fmt.Sprintf("job-%d", time.Now().UnixNano())
	// Background goroutine to process download
	file, err := processDownloadVideo(jobID, req)
	if err != nil {
		WriteError(w, "❌ Failed to process video: "+err.Error(), http.StatusInternalServerError)
		return
	}

	job := models.DownloadJob{
		JobID:       jobID,
		URL:         req.URL,
		Directory:   file.URL,
		Status:      "success",
		DirectLink:  file.URL,
		Title:       file.Title,
		Description: file.Description,
		Thumbnail:   file.Thumbnail,
		WebpageURL:  file.WebpageURL,
		Extension:   file.Ext,
		FormatID:    file.FormatID,
		FileSize:    formatSize(file.Filesize),
		Duration:    formatDuration(int64(file.Duration)),
		CreatedAt:   time.Now(),
	}

	if err := repository.SaveJob(job); err != nil {
		fmt.Printf("log.Logger: %v\n", err)
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
