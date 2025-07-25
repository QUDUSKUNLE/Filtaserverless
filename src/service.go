package src

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	urlpkg "net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

type DownloadRequest struct {
	URL       string `json:"url"`
	Quality   string `json:"quality"`
	Directory string `json:"directory"`
}

type DownloadResponse struct {
	Filename string `json:"filename"`
	Path     string `json:"path"`
	Message  string `json:"message"`
}

var (
	downloadResults = make(map[string]DownloadResponse)
	mu              = &sync.Mutex{}
)

func HandleDownload(w http.ResponseWriter, r *http.Request) {
	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, "❌ Invalid request body. Expecting JSON with 'url'", http.StatusBadRequest)
		return
	}

	// Generate a simple job ID
	jobID := fmt.Sprintf("job-%d", time.Now().UnixNano())

	// Immediately respond that job is accepted
	resp := map[string]string{
		"job_id":  jobID,
		"message": "⏳ Download started in background",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)

	// Background goroutine to process download
	go func(jobID string, req DownloadRequest) {
		log.Printf("⬇️ Starting download for job %s\n", jobID)

		filePath, err := downloadVideo(req)
		mu.Lock()
		defer mu.Unlock()

		if err != nil {
			downloadResults[jobID] = DownloadResponse{
				Message: fmt.Sprintf("❌ Download failed: %v", err),
			}
			log.Printf("❌ Job %s failed: %v\n", jobID, err)
			return
		}

		downloadResults[jobID] = DownloadResponse{
			Filename: filepath.Base(filePath),
			Path:     filePath,
			Message:  "✅ Download completed",
		}
		log.Printf("✅ Job %s completed. File: %s\n", jobID, filePath)
	}(jobID, req)
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
func downloadVideo(req DownloadRequest) (string, error) {
	// Step 1: Resolve Facebook redirect if needed
	if strings.Contains(req.URL, "facebook.com/share/r/") {
		resolved, err := resolveRedirectFully(req.URL)
		if err != nil {
			return "", fmt.Errorf("could not resolve Facebook share URL: %w", err)
		}
		req.URL = resolved
	}

	// Step 2: Ensure download directory exists
	if _, err := os.Stat(req.Directory); os.IsNotExist(err) {
		if err := os.MkdirAll(req.Directory, 0755); err != nil {
			return "", fmt.Errorf("failed to create download directory: %w", err)
		}
	}

	// Step 3: Prepare yt-dlp command
	outputTemplate := filepath.Join(req.Directory, "%(title).80s.%(ext)s")
	parsedURL, err := urlpkg.Parse(req.URL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	args := []string{
		"--restrict-filenames",
		"--merge-output-format", "mp4",
		"--no-playlist",
		"-o", outputTemplate,
		req.URL,
	}

	// Social media vs general site quality logic
	host := parsedURL.Host
	path := parsedURL.Path
	isSocial := strings.Contains(host, "facebook.com") ||
		strings.Contains(host, "fb.watch") ||
		strings.Contains(path, "/share/") ||
		strings.Contains(host, "instagram.com") ||
		strings.Contains(host, "tiktok.com")

	if isSocial {
		args = append([]string{"-f", "best"}, args...)
	} else {
		args = append([]string{"-f", "bestvideo+bestaudio/best"}, args...)
		if req.Quality != "" {
			args = append(args, "--format-sort", fmt.Sprintf("height:%s", req.Quality))
		}
	}

	cmd := exec.Command("yt-dlp", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &bytes.Buffer{}

	// Step 4: Capture pre-download state
	before, err := snapshotFiles(req.Directory)
	if err != nil {
		return "", fmt.Errorf("failed to snapshot directory: %w", err)
	}

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("yt-dlp failed: %w\nDetails: %s", err, stderr.String())
	}

	// Step 5: Detect new file
	after, err := snapshotFiles(req.Directory)
	if err != nil {
		return "", fmt.Errorf("failed to snapshot directory after download: %w", err)
	}

	newFile := diffFiles(req.Directory, before, after)
	if newFile == "" {
		return "", fmt.Errorf("yt-dlp did not produce a new downloadable file")
	}

	return newFile, nil
}

func resolveRedirectFully(shortURL string) (string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Stop after the first redirect to capture the resolved location
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
		Timeout: 15 * time.Second,
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

func snapshotFiles(dir string) (map[string]time.Time, error) {
	files := make(map[string]time.Time)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, f := range entries {
		if f.IsDir() || f.Name() == ".DS_Store" {
			continue
		}
		info, err := f.Info()
		if err != nil {
			continue
		}
		files[f.Name()] = info.ModTime()
	}
	return files, nil
}

func diffFiles(directory string, before, after map[string]time.Time) string {
	for name, modTime := range after {
		if _, exists := before[name]; !exists {
			return filepath.Join(directory, name)
		}
		if before[name] != modTime {
			return filepath.Join(directory, name)
		}
	}
	return ""
}
