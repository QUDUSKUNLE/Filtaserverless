package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	urlpkg "net/url"
	"os/exec"
	"strings"
	"time"
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

// VideoMetadata holds relevant metadata fields (customize as needed)
type VideoMetadata struct {
	Title       string  `json:"title"`
	Duration    float64 `json:"duration"`
	Thumbnail   string  `json:"thumbnail"`
	UploadDate  string  `json:"upload_date"`
	Uploader    string  `json:"uploader"`
	Description string  `json:"description"`
	WebpageURL  string  `json:"webpage_url"`
	FormatID    string  `json:"format_id"`
	Ext         string  `json:"ext"`
	Filesize    int64   `json:"filesize,omitempty"`
	URL         string  `json:"url"` // Direct download URL
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

func formatDuration(seconds int64) string {
	min := seconds / 60
	sec := seconds % 60
	return fmt.Sprintf("%d:%02d", min, sec)
}
