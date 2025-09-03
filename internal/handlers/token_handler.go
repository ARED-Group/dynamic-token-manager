package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ARED-Group/dynamic-token-manager/internal/config"
)

type TokenHandler struct {
	config *config.Config
}

type TokenRequest struct {
	Subject  string            `json:"subject"`
	Audience string            `json:"audience,omitempty"`
	Claims   map[string]interface{} `json:"claims,omitempty"`
	TTL      *int              `json:"ttl,omitempty"` // TTL in seconds
}

type TokenResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

type ValidateRequest struct {
	Token string `json:"token"`
}

type ValidateResponse struct {
	Valid     bool                   `json:"valid"`
	Claims    map[string]interface{} `json:"claims,omitempty"`
	ExpiresAt *time.Time            `json:"expires_at,omitempty"`
	Error     string                `json:"error,omitempty"`
}

func NewTokenHandler(cfg *config.Config) *TokenHandler {
	return &TokenHandler{
		config: cfg,
	}
}

// CreateToken generates a new JWT token
func (h *TokenHandler) CreateToken(w http.ResponseWriter, r *http.Request) {
	var req TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Subject == "" {
		http.Error(w, "Subject is required", http.StatusBadRequest)
		return
	}

	// Calculate expiration
	var expiration time.Duration
	if req.TTL != nil {
		expiration = time.Duration(*req.TTL) * time.Second
	} else {
		expiration = h.config.TokenExpiration
	}

	now := time.Now()
	expiresAt := now.Add(expiration)

	// Create JWT claims
	claims := jwt.MapClaims{
		"sub": req.Subject,
		"iat": now.Unix(),
		"exp": expiresAt.Unix(),
		"jti": uuid.New().String(),
		"iss": "dynamic-token-manager",
	}

	if req.Audience != "" {
		claims["aud"] = req.Audience
	}

	// Add custom claims
	for key, value := range req.Claims {
		claims[key] = value
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(h.config.JWTSecret))
	if err != nil {
		http.Error(w, "Failed to create token", http.StatusInternalServerError)
		return
	}

	// Generate refresh token
	refreshClaims := jwt.MapClaims{
		"sub": req.Subject,
		"iat": now.Unix(),
		"exp": now.Add(h.config.RefreshTokenExpiration).Unix(),
		"jti": uuid.New().String(),
		"iss": "dynamic-token-manager",
		"type": "refresh",
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(h.config.JWTSecret))
	if err != nil {
		http.Error(w, "Failed to create refresh token", http.StatusInternalServerError)
		return
	}

	response := TokenResponse{
		Token:        tokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    expiresAt,
		TokenType:    "Bearer",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ValidateToken validates a JWT token
func (h *TokenHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	var req ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.config.JWTSecret), nil
	})

	response := ValidateResponse{}

	if err != nil {
		response.Valid = false
		response.Error = err.Error()
	} else if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		response.Valid = true
		response.Claims = claims
		
		if exp, ok := claims["exp"].(float64); ok {
			expiresAt := time.Unix(int64(exp), 0)
			response.ExpiresAt = &expiresAt
		}
	} else {
		response.Valid = false
		response.Error = "Invalid token"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RefreshToken generates a new token using a refresh token
func (h *TokenHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.config.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	// Check if it's a refresh token
	if tokenType, ok := claims["type"].(string); !ok || tokenType != "refresh" {
		http.Error(w, "Not a refresh token", http.StatusUnauthorized)
		return
	}

	subject, ok := claims["sub"].(string)
	if !ok {
		http.Error(w, "Invalid subject in refresh token", http.StatusUnauthorized)
		return
	}

	// Create new access token
	now := time.Now()
	expiresAt := now.Add(h.config.TokenExpiration)

	newClaims := jwt.MapClaims{
		"sub": subject,
		"iat": now.Unix(),
		"exp": expiresAt.Unix(),
		"jti": uuid.New().String(),
		"iss": "dynamic-token-manager",
	}

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	tokenString, err := newToken.SignedString([]byte(h.config.JWTSecret))
	if err != nil {
		http.Error(w, "Failed to create new token", http.StatusInternalServerError)
		return
	}

	response := TokenResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt,
		TokenType: "Bearer",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RevokeToken revokes a token (placeholder implementation)
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
	// Extract token from context (set by JWT middleware)
	claims, ok := r.Context().Value("claims").(jwt.MapClaims)
	if !ok {
		http.Error(w, "No token claims found", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"claims": claims,
		"valid":  true,
	})
}
