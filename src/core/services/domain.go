package services

import "go.mongodb.org/mongo-driver/bson/primitive"

type (
	DownloadRequest struct {
		URL     string `json:"url"`
		Quality string `json:"quality"`
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
	UserData struct {
		ID        primitive.ObjectID `bson:"id"`
		Username  string             `bson:"username"`
		Password  string             `bson:"password"`
		Email     string             `bson:"email"`
		FirstName string             `bson:"firstName"`
		LastName  string             `bson:"lastName"`
	}
	UserResponse struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	UserSignIn struct {
		Email    string `json:"username"`
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
