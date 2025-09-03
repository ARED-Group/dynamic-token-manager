package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ARED-Group/dynamic-token-manager/internal/config"
	"github.com/ARED-Group/dynamic-token-manager/internal/models"
	"github.com/ARED-Group/dynamic-token-manager/internal/services"
)

// TokenHandler handles token-related HTTP requests
type TokenHandler struct {
	tokenService *services.TokenService
	config       *config.Config
}

// NewTokenHandler creates a new token handler
func NewTokenHandler(cfg *config.Config) *TokenHandler {
	return &TokenHandler{
		tokenService: services.NewTokenService(cfg),
		config:       cfg,
	}
}

// CreateToken creates a new token
func (h *TokenHandler) CreateToken(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed", err)
		return
	}

	token, err := h.tokenService.CreateToken(&req)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to create token", err)
		return
	}

	h.writeSuccessResponse(w, http.StatusCreated, token)
}

// RefreshToken refreshes an existing token
func (h *TokenHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	token, err := h.tokenService.RefreshToken(req.RefreshToken)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Failed to refresh token", err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, token)
}

// ValidateToken validates a token
func (h *TokenHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	var req models.ValidateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	isValid, claims, err := h.tokenService.ValidateToken(req.Token)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Token validation failed", err)
		return
	}

	response := models.ValidateTokenResponse{
		Valid:  isValid,
		Claims: claims,
	}

	h.writeSuccessResponse(w, http.StatusOK, response)
}

// RevokeToken revokes a token
func (h *TokenHandler) RevokeToken(w http.ResponseWriter, r *http.Request) {
	var req models.RevokeTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	err := h.tokenService.RevokeToken(req.Token)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to revoke token", err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, map[string]string{
		"message": "Token revoked successfully",
	})
}

// ListTokens lists all active tokens for a user
func (h *TokenHandler) ListTokens(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "user_id parameter is required", nil)
		return
	}

	tokens, err := h.tokenService.ListTokens(userID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to list tokens", err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, map[string]interface{}{
		"tokens": tokens,
		"count":  len(tokens),
	})
}

// GetTokenInfo gets information about the current token
func (h *TokenHandler) GetTokenInfo(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization header
	token := r.Header.Get("Authorization")
	if token == "" {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Authorization header is required", nil)
		return
	}

	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	info, err := h.tokenService.GetTokenInfo(token)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "Failed to get token info", err)
		return
	}

	h.writeSuccessResponse(w, http.StatusOK, info)
}

// writeSuccessResponse writes a successful JSON response
func (h *TokenHandler) writeSuccessResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := models.APIResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}

	json.NewEncoder(w).Encode(response)
}

// writeErrorResponse writes an error JSON response
func (h *TokenHandler) writeErrorResponse(w http.ResponseWriter, status int, message string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errorMsg := message
	if err != nil && h.config.Environment == "development" {
		errorMsg = errorMsg + ": " + err.Error()
	}

	response := models.APIResponse{
		Success:   false,
		Error:     &errorMsg,
		Timestamp: time.Now().UTC(),
	}

	json.NewEncoder(w).Encode(response)
}
