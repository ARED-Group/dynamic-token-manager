package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ARED-Group/dynamic-token-manager/internal/config"
	"github.com/ARED-Group/dynamic-token-manager/internal/token"
)

type TokenHandler struct {
	manager *token.TokenManager
	config  *config.Config
}

func NewTokenHandler(cfg *config.Config) *TokenHandler {
	manager, err := token.NewTokenManager()
	if err != nil {
		// For now, we'll panic since this is a critical error
		// In production, this should be handled more gracefully
		panic("Failed to create token manager: " + err.Error())
	}
	return &TokenHandler{
		manager: manager,
		config:  cfg,
	}
}

type GenerateTokenRequest struct {
	UserID   string        `json:"user_id"`
	Role     string        `json:"role"`
	Duration time.Duration `json:"duration"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type ValidateTokenRequest struct {
	Token string `json:"token"`
}

type ValidateTokenResponse struct {
	Valid  bool   `json:"valid"`
	UserID string `json:"user_id,omitempty"`
	Role   string `json:"role,omitempty"`
}

type RefreshTokenRequest struct {
	Token    string        `json:"token"`
	Duration time.Duration `json:"duration"`
}

// CreateToken handles token generation requests
func (h *TokenHandler) CreateToken(w http.ResponseWriter, r *http.Request) {
	var req GenerateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" || req.Role == "" {
		http.Error(w, "UserID and Role are required", http.StatusBadRequest)
		return
	}

	if req.Duration == 0 {
		req.Duration = 24 * time.Hour // Default to 24 hours if not specified
	}

	token, err := h.manager.GenerateToken(req.UserID, req.Role, req.Duration)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := TokenResponse{Token: token}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ValidateToken handles token validation requests
func (h *TokenHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	var req ValidateTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	claims, err := h.manager.ValidateToken(req.Token)
	if err != nil {
		response := ValidateTokenResponse{Valid: false}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	response := ValidateTokenResponse{
		Valid:  true,
		UserID: claims.UserID,
		Role:   claims.Role,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RefreshToken handles token refresh requests
func (h *TokenHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Duration == 0 {
		req.Duration = 24 * time.Hour // Default to 24 hours if not specified
	}

	newToken, err := h.manager.RefreshToken(req.Token, req.Duration)
	if err != nil {
		http.Error(w, "Failed to refresh token", http.StatusBadRequest)
		return
	}

	response := TokenResponse{Token: newToken}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RevokeToken handles token revocation requests (placeholder implementation)
func (h *TokenHandler) RevokeToken(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement token revocation with blacklist/database
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Token revocation not yet implemented",
		"status":  "pending",
	})
}

// ListTokens lists active tokens (placeholder implementation)
func (h *TokenHandler) ListTokens(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement token listing from database
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tokens": []interface{}{},
		"message": "Token listing not yet implemented",
	})
}

// GetTokenInfo returns information about the current token
func (h *TokenHandler) GetTokenInfo(w http.ResponseWriter, r *http.Request) {
	// TODO: Extract token from context (set by JWT middleware)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Token info endpoint not yet implemented",
		"status":  "pending",
	})
}