package handler

import (
	"net/http"

	"github.com/youtubebot/src/core/services"
)

func Home(w http.ResponseWriter, r *http.Request) {
	services.Home(w, r)
}
