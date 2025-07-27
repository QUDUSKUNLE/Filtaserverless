package services

import (
	"sync"
)

var (
	downloadResults = make(map[string]DownloadResponse)
	mu              = &sync.Mutex{}
)

type (
	DownloadRequest struct {
		URL       string `json:"url"`
		Quality   string `json:"quality"`
		// Directory string `json:"directory"`
	}
	UserRequest struct {
		Username        string `json:"username"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
		Email           string `json:"email"`
		FirstName       string `json:"firstName"`
		LastName        string `json:"lastName"`
	}
	UserResponse struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	UserSignIn struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	UserResetPassword struct {
		Email string `json:"email"`
	}
	DownloadResponse struct {
		Filename string `json:"filename"`
		Path     string `json:"path"`
		Message  string `json:"message"`
	}
)
