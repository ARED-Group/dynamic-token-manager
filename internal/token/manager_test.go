package token

import (
	"os"
	"testing"
	"time"
)

func TestTokenManager(t *testing.T) {
	// Set up test environment
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET_KEY")

	// Initialize token manager
	tm, err := NewTokenManager()
	if err != nil {
		t.Fatalf("Failed to create token manager: %v", err)
	}

	// Test token generation
	userID := "test_user"
	role := "admin"
	duration := 1 * time.Hour

	token, err := tm.GenerateToken(userID, role, duration)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Fatal("Generated token is empty")
	}

	// Test token validation
	claims, err := tm.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %s, got %s", userID, claims.UserID)
	}

	if claims.Role != role {
		t.Errorf("Expected Role %s, got %s", role, claims.Role)
	}

	// Test token refresh
	newToken, err := tm.RefreshToken(token, 2*time.Hour)
	if err != nil {
		t.Fatalf("Failed to refresh token: %v", err)
	}

	if newToken == "" {
		t.Fatal("Refreshed token is empty")
	}

	// Validate refreshed token
	newClaims, err := tm.ValidateToken(newToken)
	if err != nil {
		t.Fatalf("Failed to validate refreshed token: %v", err)
	}

	if newClaims.UserID != userID {
		t.Errorf("Expected UserID %s in refreshed token, got %s", userID, newClaims.UserID)
	}

	if newClaims.Role != role {
		t.Errorf("Expected Role %s in refreshed token, got %s", role, newClaims.Role)
	}
}

func TestTokenManagerWithoutSecretKey(t *testing.T) {
	// Ensure no secret key is set
	originalKey := os.Getenv("JWT_SECRET_KEY")
	os.Unsetenv("JWT_SECRET_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("JWT_SECRET_KEY", originalKey)
		}
	}()

	// Should fail to create token manager
	_, err := NewTokenManager()
	if err == nil {
		t.Fatal("Expected error when JWT_SECRET_KEY is not set")
	}
}

func TestValidateInvalidToken(t *testing.T) {
	// Set up test environment
	os.Setenv("JWT_SECRET_KEY", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET_KEY")

	tm, err := NewTokenManager()
	if err != nil {
		t.Fatalf("Failed to create token manager: %v", err)
	}

	// Test with invalid token
	_, err = tm.ValidateToken("invalid-token")
	if err == nil {
		t.Fatal("Expected error when validating invalid token")
	}
}