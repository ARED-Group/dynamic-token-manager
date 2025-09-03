package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/ARED-Group/dynamic-token-manager/internal/handlers"
	"github.com/ARED-Group/dynamic-token-manager/internal/token"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using system environment variables")
	}

	// Initialize token manager
	tokenManager, err := token.NewTokenManager()
	if err != nil {
		log.Fatal("Failed to initialize token manager:", err)
	}

	// Initialize router
	router := mux.NewRouter()

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Set up routes
	setupRoutes(router, tokenManager)

	// Start server
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), router); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func setupRoutes(router *mux.Router, tokenManager *token.TokenManager) {
	// Initialize token handler
	tokenHandler := handlers.NewTokenHandler(tokenManager)

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Token management endpoints
	router.HandleFunc("/tokens/generate", tokenHandler.GenerateToken).Methods("POST")
	router.HandleFunc("/tokens/validate", tokenHandler.ValidateToken).Methods("POST")
	router.HandleFunc("/tokens/refresh", tokenHandler.RefreshToken).Methods("POST")
}