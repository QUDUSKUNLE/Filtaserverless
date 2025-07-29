package services

import "go.mongodb.org/mongo-driver/bson/primitive"

type (
	DownloadRequest struct {
		URL    string `json:"url" validate:"required"`
		UserID string `json:"user_id,omitempty"`
	}
	UserRequest struct {
		Username        string `json:"username,omitempty"`
		Password        string `json:"password" validate:"required"`
		ConfirmPassword string `json:"confirm_password" validate:"required"`
		Email           string `json:"email" validate:"email,required"`
		FirstName       string `json:"first_name" validate:"required"`
		LastName        string `json:"last_name" validate:"required"`
	}
	UserData struct {
		ID        primitive.ObjectID `bson:"_id,omitempty"`
		Username  string             `bson:"username"`
		Password  string             `bson:"password"`
		Email     string             `bson:"email"`
		FirstName string             `bson:"first_name"`
		LastName  string             `bson:"last_name"`
	}
	UserResponse struct {
		ID      string `json:"id"`
		Message string `json:"message"`
	}
	UserSignIn struct {
		Email    string `json:"email" validate:"email,required"`
		Password string `json:"password" validate:"required"`
	}
	UserResetPassword struct {
		Email string `json:"email" validate:"email,required"`
	}
	DownloadResponse struct {
		Filename string `json:"filename"`
		Path     string `json:"path"`
		Message  string `json:"message"`
	}
	// VideoMetadata holds relevant metadata fields (customize as needed)
	VideoMetadata struct {
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
)
