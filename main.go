package handler

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	chiadapter "github.com/awslabs/aws-lambda-go-api-proxy/chi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/youtubebot/src/adapters/db"
	middle "github.com/youtubebot/src/adapters/middleware"
	"github.com/youtubebot/src/core/services"
)

var chiLambda *chiadapter.ChiLambda

func init() {
	_ = godotenv.Load()
	db.Connect()
	router := setupRouter()
	chiLambda = chiadapter.New(router)
}

func setupRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middle.CorsMiddleware)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", services.Home)
	r.Post("/analyse", services.HandleDownload)
	r.Get("/status/{jobID}", services.GetDownloadStatus)

	return r
}

func main() {
	if os.Getenv("LOCAL") == "true" {
		startLocalServer()
	}
	lambda.Start(chiLambda.ProxyWithContext)
}

func startLocalServer() {
	fmt.Println("🧪 Running locally on :9096")
	log.Fatal(http.ListenAndServe(":9096", setupRouter()))
}
