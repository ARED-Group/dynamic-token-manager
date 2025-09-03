package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ARED-Group/dynamic-token-manager/internal/models"
	"github.com/ARED-Group/dynamic-token-manager/internal/services"
)

type TokenHandler struct {
	tokenService  *services.TokenService
	deviceService *services.DeviceService
}

func NewTokenHandler(tokenService *services.TokenService, deviceService *services.DeviceService) *TokenHandler {
	return &TokenHandler{
		tokenService:  tokenService,
		deviceService: deviceService,
	}
}

// GenerateToken handles token generation requests
func (h *TokenHandler) GenerateToken(w http.ResponseWriter, r *http.Request) {
	var req models.TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get device serial from context (set by middleware)
	if deviceSerial, ok := r.Context().Value("device_serial").(string); ok {
		req.DeviceSerial = deviceSerial
	}

	token, err := h.tokenService.GenerateToken(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(token)
}

// ValidateToken handles token validation requests
func (h *TokenHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	claims, err := h.tokenService.ValidateToken(req.Token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":  true,
		"claims": claims,
	})
}

// RefreshToken handles token refresh requests
func (h *TokenHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	newToken, err := h.tokenService.RefreshToken(req.Token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newToken)
}
