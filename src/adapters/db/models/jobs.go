package models

import "time"

type DownloadJob struct {
	JobID     string    `bson:"job_id"`
	URL       string    `bson:"url"`
	Directory string    `bson:"directory"`
	Status    string    `bson:"status"` // pending, success, failed
	CreatedAt time.Time `bson:"created_at"`
}
