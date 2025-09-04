package services

import (
	"fmt"
	"time"

	"github.com/ARED-Group/dynamic-token-manager/internal/config"
	"github.com/ARED-Group/dynamic-token-manager/internal/github"
	"github.com/ARED-Group/dynamic-token-manager/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

type TokenService struct {
	config    *config.Config
	githubApp *github.App
}

func NewTokenService(cfg *config.Config) (*TokenService, error) {
	var githubApp *github.App
	var err error

	if cfg.GitHubAppID != "" {
		githubApp, err = github.NewApp(cfg.GitHubAppID, cfg.GitHubInstallationID, cfg.GitHubPrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create GitHub app: %w", err)
		}
	}

	return &TokenService{
		config:    cfg,
		githubApp: githubApp,
	}, nil
}

// GenerateToken creates a new JWT token
func (s *TokenService) GenerateToken(req *models.TokenRequest) (*models.TokenResponse, error) {
	now := time.Now()
	expiry := now.Add(s.config.TokenExpiration)

	claims := jwt.MapClaims{
		"device_serial": req.DeviceSerial,
		"token_type":    req.TokenType,
		"scopes":        req.Scopes,
		"iat":           now.Unix(),
		"exp":           expiry.Unix(),
		"iss":           "dynamic-token-manager",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return &models.TokenResponse{
		Token:     tokenString,
		ExpiresAt: expiry,
		TokenType: req.TokenType,
		Scopes:    req.Scopes,
	}, nil
}

// ValidateToken validates a JWT token
func (s *TokenService) ValidateToken(tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GetGitHubRegistryToken gets a GitHub container registry token
func (s *TokenService) GetGitHubRegistryToken(req *models.GitHubRegistryTokenRequest) (*models.GitHubRegistryTokenResponse, error) {
	if s.githubApp == nil {
		return nil, fmt.Errorf("GitHub App not configured")
	}

	// Get GitHub installation token
	githubToken, err := s.githubApp.GetInstallationToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get GitHub token: %w", err)
	}

	return &models.GitHubRegistryTokenResponse{
		Token:     githubToken.Token,
		ExpiresAt: githubToken.ExpiresAt,
		Registry:  s.config.RegistryURL,
		Username:  s.config.RegistryUsername,
	}, nil
}

// RefreshToken refreshes an existing token
func (s *TokenService) RefreshToken(tokenString string) (*models.TokenResponse, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token for refresh: %w", err)
	}

	// Extract device serial from existing token
	deviceSerial, ok := (*claims)["device_serial"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid device serial in token")
	}

	// Extract scopes
	scopesInterface, ok := (*claims)["scopes"]
	var scopes []string
	if ok && scopesInterface != nil {
		if scopesList, ok := scopesInterface.([]interface{}); ok {
			for _, scope := range scopesList {
				if scopeStr, ok := scope.(string); ok {
					scopes = append(scopes, scopeStr)
				}
			}
		}
	}

	// Generate new token
	return s.GenerateToken(&models.TokenRequest{
		DeviceSerial: deviceSerial,
		TokenType:    "refresh",
		Scopes:       scopes,
	})
}
