package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	urlpkg "net/url"
	"os/exec"
	"strings"
	"time"

	"github.com/youtubebot/src/adapters/db"
	"go.mongodb.org/mongo-driver/bson"
)

func resolveRedirectFully(shortURL string) (string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Stop after the first redirect to capture the resolved location
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", shortURL, nil)
	if err != nil {
		return "", err
	}

	// Set a real user-agent to avoid bot filtering
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	finalURL := resp.Request.URL.String()
	return finalURL, nil
}

// getDirectDownloadURL determines the platform (YouTube, Facebook, Instagram),
// normalizes share/redirect URLs, and invokes yt-dlp to extract direct media metadata.
func getDirectDownloadURL(rawURL string) (*VideoMetadata, error) {
	// Normalize input
	normalized := strings.TrimSpace(rawURL)

	// Facebook: resolve short share/fb.watch redirects
	if strings.Contains(normalized, "facebook.com/share/r/") || strings.Contains(normalized, "fb.watch/") {
		resolved, err := resolveRedirectFully(normalized)
		if err != nil {
			return nil, fmt.Errorf("could not resolve Facebook share URL: %w", err)
		}
		normalized = resolved
	}

	// Instagram: strip tracking query parameters like igshid to stabilize extraction
	if strings.Contains(normalized, "instagram.com") {
		if u, err := urlpkg.Parse(normalized); err == nil {
			if q := u.Query(); len(q) > 0 {
				// remove common tracking params without risking required identifiers
				q.Del("igshid")
				q.Del("utm_source")
				q.Del("utm_medium")
				q.Del("utm_campaign")
				u.RawQuery = q.Encode()
			}
			normalized = u.String()
		}
	}

	// Validate URL
	parsedURL, err := urlpkg.Parse(normalized)
	if err != nil || !parsedURL.IsAbs() {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	lowerHost := strings.ToLower(parsedURL.Hostname())

	// Build yt-dlp args per provider
	args := []string{"-j", "--simulate"}

	switch {
	case strings.Contains(lowerHost, "youtube.com") || strings.Contains(lowerHost, "youtu.be"):
		// Prefer best video+audio merged if available, fallback to best single stream
		args = append(args, "-f", "best[ext=mp4]/best")
	case strings.Contains(lowerHost, "facebook.com"):
		// Facebook can be stricter; prefer a robust single best with retries
		args = append(args, "-f", "b", "--force-ipv4", "--retries", "3")
	case strings.Contains(lowerHost, "instagram.com"):
		// Instagram reels/posts
		args = append(args, "-f", "b", "--retries", "3")
	default:
		// Generic fallback
		args = append(args, "-f", "best")
	}

	args = append(args, normalized)
	cmd := exec.Command("yt-dlp", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("yt-dlp failed: %w\nDetails: %s", err, stderr.String())
	}

	var meta VideoMetadata
	if err := json.Unmarshal(stdout.Bytes(), &meta); err != nil {
		return nil, fmt.Errorf("failed to parse yt-dlp JSON: %w", err)
	}

	return &meta, nil
}

func processDownloadVideo(jobID string, req DownloadRequest) (*VideoMetadata, error) {
	log.Printf("⬇️ Starting fetch for job %s\n", jobID)

	file, err := getDirectDownloadURL(req.URL)
	if err != nil {
		log.Printf("❌ Job %s failed: %v\n", jobID, err)
		return nil, err
	}

	job := bson.M{
		// "user_id":     req.UserID,
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

func formatSize(bytes int64) string {
	return fmt.Sprintf("%.2f MB", float64(bytes)/(1024*1024))
}

func formatDuration(seconds int64) string {
	min := seconds / 60
	sec := seconds % 60
	return fmt.Sprintf("%d:%02d", min, sec)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func WriteError(w http.ResponseWriter, message string, status int) {
	writeJSON(w, status, map[string]string{"error": message})
}

// Extract userID from request context
func GetUserID(r *http.Request) string {
	if id, ok := r.Context().Value("ID").(string); ok {
		return id
	}
	return ""
}
