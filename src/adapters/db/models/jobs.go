package models

import "time"

type DownloadJob struct {
	JobID       string    `bson:"job_id"`
	URL         string    `bson:"url"`
	Directory   string    `bson:"directory"`
	Status      string    `bson:"status"`      // pending, success, failed
	DirectLink  string    `bson:"direct_link"` // direct link to the downloaded file
	Title       string    `bson:"title"`
	Description string    `bson:"description"`
	Thumbnail   string    `bson:"thumbnail"`
	WebpageURL  string    `bson:"webpage_url"`
	Extension   string    `bson:"extension"`
	FormatID    string    `bson:"format_id"`
	FileSize    string    `bson:"filesize"`
	Duration    string    `bson:"duration"`
	CreatedAt   time.Time `bson:"created_at"`
}
