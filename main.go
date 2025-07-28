package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/youtubebot/src/adapters/db"
	middle "github.com/youtubebot/src/adapters/middleware"
	"github.com/youtubebot/src/core/services"
)

func init() {
	_ = godotenv.Load()
	db.Connect()
}

func setupRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middle.CorsMiddleware)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", services.Home)
	r.Post("/analyse", services.Analyse)
	r.Get("/status/{jobID}", services.GetStatus)
	r.Post("/login", services.Login)
	r.Post("/register", services.SignUp)
	r.Post("/subscribe", services.Subscribe)

	return r
}

func main() {
	fmt.Println("🧪 Running locally on :9096")
	log.Fatal(http.ListenAndServe(":9096", setupRouter()))
}
