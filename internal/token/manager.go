package token

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenManager handles JWT token operations
type TokenManager struct {
	secretKey []byte
}

// Claims represents the claims in our JWT token
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// NewTokenManager creates a new instance of TokenManager
func NewTokenManager() (*TokenManager, error) {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		return nil, errors.New("JWT_SECRET_KEY not set in environment")
	}
	return &TokenManager{
		secretKey: []byte(secretKey),
	}, nil
}

// GenerateToken creates a new JWT token for a user
func (tm *TokenManager) GenerateToken(userID, role string, expirationTime time.Duration) (string, error) {
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expirationTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.secretKey)
}

// ValidateToken verifies if a token is valid and returns its claims
func (tm *TokenManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tm.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken generates a new token while preserving the original claims
func (tm *TokenManager) RefreshToken(oldTokenString string, newExpirationTime time.Duration) (string, error) {
	claims, err := tm.ValidateToken(oldTokenString)
	if err != nil {
		return "", err
	}

	// Create new token with same claims but new expiration
	return tm.GenerateToken(claims.UserID, claims.Role, newExpirationTime)
}