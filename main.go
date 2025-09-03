package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/ARED-Group/dynamic-token-manager/api"
	"github.com/ARED-Group/dynamic-token-manager/internal/config"
)

func main() {
	// Load configuration
	cfg := config.Load()
	
	// Validate GitHub configuration if GitHub App is enabled
	if cfg.GitHubAppID != "" {
		if err := cfg.ValidateGitHubConfig(); err != nil {
			log.Printf("Warning: GitHub configuration validation failed: %v", err)
			log.Printf("GitHub functionality will be limited")
		} else {
			log.Printf("GitHub App configured successfully (App ID: %s)", cfg.GitHubAppID)
		}
	}
	
	// Create router
	router := mux.NewRouter()
	
	// Setup routes
	if err := api.SetupRoutes(router, cfg); err != nil {
		log.Fatalf("Failed to setup routes: %v", err)
	}
	
	// Create server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.ServerReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.ServerWriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.ServerIdleTimeout) * time.Second,
	}
	
	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on port %s", cfg.Port)
		log.Printf("Environment: %s", cfg.Environment)
		log.Printf("GitHub App ID: %s", cfg.GitHubAppID)
		log.Printf("Device Auth Enabled: %t", cfg.DeviceAuthEnabled)
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()
	
	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	
	// Give outstanding requests a 30-second deadline to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	
	log.Println("Server exited")
}
