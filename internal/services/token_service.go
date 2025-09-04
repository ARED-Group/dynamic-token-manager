package services

import (
	"context"
	"sync"
	"time"

	"github.com/ARED-Group/dynamic-token-manager/config"
	"github.com/ARED-Group/dynamic-token-manager/internal/github"
)

// TokenService manages GitHub installation tokens with a simple in-memory cache.
type TokenService struct {
	cfg *config.Config
	app *github.App

	mu          sync.Mutex
	cachedToken string
	expiry      time.Time
}

// NewTokenService constructs a new TokenService.
func NewTokenService(cfg *config.Config, app *github.App) *TokenService {
	return &TokenService{
		cfg: cfg,
		app: app,
	}
}

// GetToken returns a cached token if valid, otherwise fetches a fresh installation token.
func (s *TokenService) GetToken(ctx context.Context) (string, time.Time, error) {
	s.mu.Lock()
	token := s.cachedToken
	exp := s.expiry
	s.mu.Unlock()

	// If token exists and is not about to expire, return it.
	if token != "" && time.Until(exp) > 1*time.Minute {
		return token, exp, nil
	}

	// Otherwise fetch a new token from GitHub App.
	tr, err := s.app.GetInstallationToken()
	if err != nil {
		return "", time.Time{}, err
	}

	var newExpiry time.Time
	if !tr.ExpiresAt.IsZero() {
		newExpiry = tr.ExpiresAt
	} else if s.cfg != nil && s.cfg.TokenExpiration > 0 {
		newExpiry = time.Now().Add(s.cfg.TokenExpiration)
	} else {
		newExpiry = time.Now().Add(50 * time.Minute)
	}

	s.mu.Lock()
	s.cachedToken = tr.Token
	s.expiry = newExpiry
	s.mu.Unlock()

	return tr.Token, newExpiry, nil
}

// Invalidate clears the cached token.
func (s *TokenService) Invalidate() {
	s.mu.Lock()
	s.cachedToken = ""
	s.expiry = time.Time{}
	s.mu.Unlock()
}