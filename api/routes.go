package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/ARED-Group/dynamic-token-manager/internal/config"
	"github.com/ARED-Group/dynamic-token-manager/internal/handlers"
	"github.com/ARED-Group/dynamic-token-manager/internal/middleware"
	"github.com/ARED-Group/dynamic-token-manager/internal/services"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *mux.Router, cfg *config.Config) error {
	// Initialize services
	tokenService, err := services.NewTokenService(cfg)
	if err != nil {
		return err
	}
	deviceService := services.NewDeviceService(cfg)

	// Initialize handlers
	tokenHandler := handlers.NewTokenHandler(tokenService, deviceService)
	githubHandler := handlers.NewGitHubRegistryHandler(cfg, tokenService, deviceService)
	healthHandler := handlers.NewHealthHandler()
	
	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg, tokenService, deviceService)
	rateLimiter := middleware.NewRateLimiter(time.Minute, time.Hour)
	
	// Global middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logging())
	router.Use(rateLimiter.RateLimitMiddleware(cfg.RateLimitPerMinute))
	router.Use(middleware.RequestID())
	
	// Health check endpoints (no auth required)
	router.HandleFunc("/health", healthHandler.Health).Methods("GET")
	router.HandleFunc("/ready", healthHandler.Ready).Methods("GET")
	
	// API version prefix
	api := router.PathPrefix("/api/v1").Subrouter()
	
	// GitHub status endpoint (no auth required for monitoring)
	api.HandleFunc("/github/status", githubHandler.GetGitHubStatus).Methods("GET")
	
	// Token management endpoints (require device auth)
	tokenRoutes := api.PathPrefix("/tokens").Subrouter()
	tokenRoutes.Use(authMiddleware.DeviceAuthMiddleware)
	tokenRoutes.HandleFunc("", tokenHandler.GenerateToken).Methods("POST")
	tokenRoutes.HandleFunc("/refresh", tokenHandler.RefreshToken).Methods("POST")
	tokenRoutes.HandleFunc("/validate", tokenHandler.ValidateToken).Methods("POST")
	
	// GitHub Registry endpoints (require device auth) - CRITICAL FOR sync_containers.py
	githubRoutes := api.PathPrefix("/github").Subrouter()
	githubRoutes.Use(authMiddleware.DeviceAuthMiddleware)
	
	// Main endpoint your Python script needs
	githubRoutes.HandleFunc("/registry-token", githubHandler.GetRegistryToken).Methods("GET")
	githubRoutes.HandleFunc("/registry-credentials", githubHandler.GetRegistryCredentials).Methods("GET")
	githubRoutes.HandleFunc("/token/refresh", githubHandler.RefreshGitHubToken).Methods("POST")
	githubRoutes.HandleFunc("/token/validate", githubHandler.ValidateGitHubToken).Methods("POST")
	
	// Protected endpoints (require JWT authentication)
	protected := api.PathPrefix("/").Subrouter()
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))
	protected.HandleFunc("/tokens/info", tokenHandler.GetTokenInfo).Methods("GET")
	
	// Metrics endpoint (if enabled)
	if cfg.EnableMetrics {
		router.HandleFunc("/metrics", metricsHandler).Methods("GET")
	}
	
	// 404 handler
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	
	return nil
}

// metricsHandler serves Prometheus metrics
func metricsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement Prometheus metrics
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("# Metrics endpoint - TODO: Implement Prometheus metrics\n"))
}

// notFoundHandler returns a JSON 404 response
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	
	response := map[string]interface{}{
		"error":     "Not Found",
		"message":   "The requested endpoint does not exist",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"path":      r.URL.Path,
	}
	
	json.NewEncoder(w).Encode(response)
}
