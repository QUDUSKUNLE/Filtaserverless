package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func Analyse(w http.ResponseWriter, r *http.Request) {
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
