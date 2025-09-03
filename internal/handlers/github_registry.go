package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ARED-Group/dynamic-token-manager/internal/config"
	"github.com/ARED-Group/dynamic-token-manager/internal/models"
	"github.com/ARED-Group/dynamic-token-manager/internal/services"
)

type GitHubRegistryHandler struct {
	config        *config.Config
	tokenService  *services.TokenService
	deviceService *services.DeviceService
}

func NewGitHubRegistryHandler(cfg *config.Config, tokenService *services.TokenService, deviceService *services.DeviceService) *GitHubRegistryHandler {
	return &GitHubRegistryHandler{
		config:        cfg,
		tokenService:  tokenService,
		deviceService: deviceService,
	}
}

// GetRegistryToken - This is the endpoint your sync_containers.py will call
func (h *GitHubRegistryHandler) GetRegistryToken(w http.ResponseWriter, r *http.Request) {
	// Get device serial from context (set by middleware)
	deviceSerial, ok := r.Context().Value("device_serial").(string)
	if !ok {
		log.Printf("No device serial found in request context")
		h.sendErrorResponse(w, "Device authentication required", http.StatusUnauthorized)
		return
	}

	log.Printf("GitHub registry token requested by device: %s", deviceSerial)

	// Validate GitHub App configuration
	if err := h.config.ValidateGitHubConfig(); err != nil {
		log.Printf("GitHub configuration error: %v", err)
		h.sendErrorResponse(w, "GitHub integration not configured", http.StatusServiceUnavailable)
		return
	}

	// Create request for GitHub registry token
	req := &models.GitHubRegistryTokenRequest{
		DeviceSerial: deviceSerial,
		Repository:   "ared-group", // Your organization
	}

	// Get GitHub token from service
	token, err := h.tokenService.GetGitHubRegistryToken(req)
	if err != nil {
		log.Printf("Failed to get GitHub registry token for device %s: %v", deviceSerial, err)
		h.sendErrorResponse(w, "Failed to obtain GitHub token", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully generated GitHub registry token for device: %s (expires: %v)", deviceSerial, token.ExpiresAt)

	// Return token response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(token)
}

// GetRegistryCredentials - Alternative endpoint that returns ready-to-use credentials
func (h *GitHubRegistryHandler) GetRegistryCredentials(w http.ResponseWriter, r *http.Request) {
	deviceSerial, ok := r.Context().Value("device_serial").(string)
	if !ok {
		h.sendErrorResponse(w, "Device authentication required", http.StatusUnauthorized)
		return
	}

	// Get GitHub token
	req := &models.GitHubRegistryTokenRequest{
		DeviceSerial: deviceSerial,
		Repository:   "ared-group",
	}

	token, err := h.tokenService.GetGitHubRegistryToken(req)
	if err != nil {
		log.Printf("Failed to get GitHub registry credentials for device %s: %v", deviceSerial, err)
		h.sendErrorResponse(w, "Failed to obtain registry credentials", http.StatusInternalServerError)
		return
	}

	// Return credentials in format ready for balena login
	credentials := map[string]interface{}{
		"registry": token.Registry,
		"username": token.Username,
		"token":    token.Token,
		"expires_at": token.ExpiresAt,
		"login_command": fmt.Sprintf("echo %s | balena login %s -u %s --password-stdin", 
			token.Token, token.Registry, token.Username),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(credentials)
}

// RefreshGitHubToken - Force refresh GitHub token
func (h *GitHubRegistryHandler) RefreshGitHubToken(w http.ResponseWriter, r *http.Request) {
	deviceSerial, ok := r.Context().Value("device_serial").(string)
	if !ok {
		h.sendErrorResponse(w, "Device authentication required", http.StatusUnauthorized)
		return
	}

	log.Printf("Force refresh GitHub token requested by device: %s", deviceSerial)

	// TODO: Implement force refresh logic
	// For now, just get a new token (GitHub App tokens are always fresh)
	req := &models.GitHubRegistryTokenRequest{
		DeviceSerial: deviceSerial,
		Repository:   "ared-group",
	}

	token, err := h.tokenService.GetGitHubRegistryToken(req)
	if err != nil {
		log.Printf("Failed to refresh GitHub token for device %s: %v", deviceSerial, err)
		h.sendErrorResponse(w, "Failed to refresh GitHub token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(token)
}

// ValidateGitHubToken - Validate if GitHub token is still valid
func (h *GitHubRegistryHandler) ValidateGitHubToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual GitHub token validation
	// For now, return a simple response
	response := map[string]interface{}{
		"valid": true,
		"message": "Token validation not implemented yet",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetGitHubStatus - Health check for GitHub App integration
func (h *GitHubRegistryHandler) GetGitHubStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"github_app_configured": h.config.GitHubAppID != "",
		"github_app_id": h.config.GitHubAppID,
		"installation_id": h.config.GitHubInstallationID,
		"private_key_configured": h.config.GitHubPrivateKeyPath != "",
		"registry_url": h.config.RegistryURL,
		"registry_username": h.config.RegistryUsername,
	}

	// Try to validate GitHub configuration
	if err := h.config.ValidateGitHubConfig(); err != nil {
		status["error"] = err.Error()
		status["healthy"] = false
	} else {
		status["healthy"] = true
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

// Helper method to send error responses
func (h *GitHubRegistryHandler) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	errorResp := models.ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
		Code:    statusCode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResp)
}
