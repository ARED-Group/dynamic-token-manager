package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/ARED-Group/dynamic-token-manager/internal/token"
)

func TestTokenHandlerGenerateToken(t *testing.T) {
	// Set up test environment
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET_KEY")

	// Create token manager and handler
	tm, err := token.NewTokenManager()
	if err != nil {
		t.Fatalf("Failed to create token manager: %v", err)
	}

	handler := NewTokenHandler(tm)

	// Create test request
	reqBody := GenerateTokenRequest{
		UserID:   "test_user",
		Role:     "admin",
		Duration: 1 * time.Hour,
	}
	jsonData, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", "/tokens/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.GenerateToken(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response TokenResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Token == "" {
		t.Error("Expected token in response, got empty string")
	}
}

func TestTokenHandlerValidateToken(t *testing.T) {
	// Set up test environment
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET_KEY")

	// Create token manager and handler
	tm, err := token.NewTokenManager()
	if err != nil {
		t.Fatalf("Failed to create token manager: %v", err)
	}

	handler := NewTokenHandler(tm)

	// Generate a test token first
	testToken, err := tm.GenerateToken("test_user", "admin", 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	// Create test request
	reqBody := ValidateTokenRequest{
		Token: testToken,
	}
	jsonData, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", "/tokens/validate", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.ValidateToken(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response ValidateTokenResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Valid {
		t.Error("Expected token to be valid")
	}

	if response.UserID != "test_user" {
		t.Errorf("Expected UserID 'test_user', got '%s'", response.UserID)
	}

	if response.Role != "admin" {
		t.Errorf("Expected Role 'admin', got '%s'", response.Role)
	}
}

func TestTokenHandlerGenerateTokenBadRequest(t *testing.T) {
	// Set up test environment
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET_KEY")

	// Create token manager and handler
	tm, err := token.NewTokenManager()
	if err != nil {
		t.Fatalf("Failed to create token manager: %v", err)
	}

	handler := NewTokenHandler(tm)

	// Create test request with missing UserID
	reqBody := GenerateTokenRequest{
		UserID:   "",
		Role:     "admin",
		Duration: 1 * time.Hour,
	}
	jsonData, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", "/tokens/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	handler.GenerateToken(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}