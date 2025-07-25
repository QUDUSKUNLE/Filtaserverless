package src

import (
	"sync"
)

var (
	downloadResults = make(map[string]DownloadResponse)
	mu              = &sync.Mutex{}
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
