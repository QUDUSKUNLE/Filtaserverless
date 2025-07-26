package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	urlpkg "net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func DownloadVideo(req DownloadRequest) (string, error) {
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
	outputTemplate := filepath.Join(req.Directory, "%(title).80s.%(id)s.%(ext)s")

	parsedURL, err := urlpkg.Parse(req.URL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	args := []string{
		"--restrict-filenames",
		"--merge-output-format",
		"mp4",
		"--no-playlist",
		"--no-part", // <--- avoids .part temp files
		"--downloader",
		"ffmpeg", // <--- force use of ffmpeg
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

// VideoMetadata holds relevant metadata fields (customize as needed)
type VideoMetadata struct {
	Title       string `json:"title"`
	Duration    int    `json:"duration"`
	Thumbnail   string `json:"thumbnail"`
	UploadDate  string `json:"upload_date"`
	Uploader    string `json:"uploader"`
	Description string `json:"description"`
	WebpageURL  string `json:"webpage_url"`
	FormatID    string `json:"format_id"`
	Ext         string `json:"ext"`
	Filesize    int64  `json:"filesize,omitempty"`
	URL         string `json:"url"` // Direct download URL
}

// GetDirectDownloadURL retrieves both direct download link and video metadata using yt-dlp.
func getDirectDownloadURL(rawURL string) (*VideoMetadata, error) {
	// Resolve Facebook redirect if needed
	if strings.Contains(rawURL, "facebook.com/share/r/") {
		resolved, err := resolveRedirectFully(rawURL)
		if err != nil {
			return nil, fmt.Errorf("could not resolve Facebook share URL: %w", err)
		}
		rawURL = resolved
	}

	// Validate URL
	parsedURL, err := urlpkg.Parse(rawURL)
	if err != nil || !parsedURL.IsAbs() {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// yt-dlp -f best -j --simulate [URL]
	args := []string{"-f", "best", "-j", "--simulate", rawURL}
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

func formatSize(bytes int64) string {
	return fmt.Sprintf("%.2f MB", float64(bytes)/(1024*1024))
}

func formatDuration(seconds int) string {
	min := seconds / 60
	sec := seconds % 60
	return fmt.Sprintf("%d:%02d", min, sec)
}
