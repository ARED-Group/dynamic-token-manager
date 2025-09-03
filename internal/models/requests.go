package models

import "time"

// TokenRequest represents a request for a new token
type TokenRequest struct {
	DeviceSerial string   `json:"device_serial,omitempty"`
	TokenType    string   `json:"token_type,omitempty"`
	Scopes       []string `json:"scopes,omitempty"`
}

// TokenResponse represents a token response
type TokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	TokenType string    `json:"token_type"`
	Scopes    []string  `json:"scopes,omitempty"`
}

// GitHubRegistryTokenRequest represents a request for GitHub container registry token
type GitHubRegistryTokenRequest struct {
	DeviceSerial string `json:"device_serial"`
	Repository   string `json:"repository,omitempty"`
}

// GitHubRegistryTokenResponse represents GitHub registry token response
type GitHubRegistryTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Registry  string    `json:"registry"`
	Username  string    `json:"username"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// DeviceValidationRequest for device authentication
type DeviceValidationRequest struct {
	SerialNumber string `json:"serial_number"`
	Signature    string `json:"signature,omitempty"`
}

// DeviceValidationResponse for device authentication
type DeviceValidationResponse struct {
	Valid      bool   `json:"valid"`
	DeviceID   string `json:"device_id,omitempty"`
	Message    string `json:"message,omitempty"`
}
