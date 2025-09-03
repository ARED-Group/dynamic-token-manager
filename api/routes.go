package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/ARED-Group/dynamic-token-manager/internal/config"
	"github.com/ARED-Group/dynamic-token-manager/internal/handlers"
	"github.com/ARED-Group/dynamic-token-manager/internal/middleware"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *mux.Router, cfg *config.Config) {
	// Initialize handlers
	tokenHandler := handlers.NewTokenHandler(cfg)
	healthHandler := handlers.NewHealthHandler()
	
	// API version prefix
	api := router.PathPrefix("/api/v1").Subrouter()
	
	// Middleware
	api.Use(middleware.CORS())
	api.Use(middleware.Logging())
	api.Use(middleware.RateLimit(cfg.RateLimitPerMinute))
	api.Use(middleware.RequestID())
	
	// Health check endpoints
	router.HandleFunc("/health", healthHandler.Health).Methods("GET")
	router.HandleFunc("/ready", healthHandler.Ready).Methods("GET")
	
	// Token management endpoints
	api.HandleFunc("/tokens", tokenHandler.CreateToken).Methods("POST")
	api.HandleFunc("/tokens/refresh", tokenHandler.RefreshToken).Methods("POST")
	api.HandleFunc("/tokens/validate", tokenHandler.ValidateToken).Methods("POST")
	api.HandleFunc("/tokens/revoke", tokenHandler.RevokeToken).Methods("POST")
	api.HandleFunc("/tokens", tokenHandler.ListTokens).Methods("GET")
	
	// Token info endpoint (requires authentication)
	protected := api.PathPrefix("/").Subrouter()
	protected.Use(middleware.JWTAuth(cfg.JWTSecret))
	protected.HandleFunc("/tokens/info", tokenHandler.GetTokenInfo).Methods("GET")
	
	// Metrics endpoint (if enabled)
	if cfg.EnableMetrics {
		router.HandleFunc("/metrics", metricsHandler).Methods("GET")
	}
	
	// 404 handler
	router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
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
