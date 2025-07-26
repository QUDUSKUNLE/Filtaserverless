package repository


import (
	"context"
	"time"

	"github.com/youtubebot/src/adapters/db"
	"github.com/youtubebot/src/adapters/db/models"
)

func SaveJob(job models.DownloadJob) error {
	collection := db.MongoDB.Collection("jobs")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, job)
	return err
}
