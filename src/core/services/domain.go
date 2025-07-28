package services

import "go.mongodb.org/mongo-driver/bson/primitive"

type (
	DownloadRequest struct {
		URL     string `json:"url"`
		Quality string `json:"quality"`
	}
	UserRequest struct {
		Username        string `json:"username,omitempty"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirm_password"`
		Email           string `json:"email"`
		FirstName       string `json:"first_name"`
		LastName        string `json:"last_name"`
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
		ID      string `json:"id"`
		Message string `json:"message"`
	}
	UserSignIn struct {
		Email    string `json:"email"`
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
