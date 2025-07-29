package services

import (
	"context"
	"net/http"
	"time"

	"github.com/youtubebot/src/adapters/db"
	"github.com/youtubebot/src/adapters/db/models"
	"gopkg.in/mgo.v2/bson"
)

func GetStatus(w http.ResponseWriter, r *http.Request) {
	jobID := r.URL.Query().Get("jobID")
	collection := db.MongoDB.Collection("jobs")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var job models.DownloadJob
	err := collection.FindOne(ctx, bson.M{"job_id": jobID}).Decode(&job)
	if err != nil {
		WriteError(w, "Job not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, job)
}
