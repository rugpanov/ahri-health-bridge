package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"

	"github.com/rugpanov/ahri-health-bridge/controllers"
	"github.com/rugpanov/ahri-health-bridge/gateways"
	"github.com/rugpanov/ahri-health-bridge/handlers"
	"github.com/rugpanov/ahri-health-bridge/utils"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("error loading .env file: %v", err)
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logFile := os.Getenv("LOG_FILE")
	if logFile == "" {
		logFile = "payloads.json"
	}

	logger, err := gateways.NewLogger(logFile)
	if err != nil {
		log.Fatalf("error creating logger: %v", err)
	}

	stepsCtrl := controllers.NewStepsController(logger)
	stepsHandler := handlers.NewStepsHandler(stepsCtrl)

	r := chi.NewRouter()
	r.Use(utils.APIKeyMiddleware(apiKey))
	r.Post("/health/steps", stepsHandler.ServeHTTP)

	fmt.Printf("ahri-health-bridge listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
