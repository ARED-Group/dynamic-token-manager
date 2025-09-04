package github

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// TokenResponse is a simple representation of the installation token returned from GitHub.
type TokenResponse struct {
	Token     string
	ExpiresAt time.Time
}

// App represents a GitHub App instance used to fetch installation tokens.
type App struct {
	AppID          string
	InstallationID string
	PrivateKey     *rsa.PrivateKey
}

// LoadPrivateKey loads the private key from a file.
func LoadPrivateKey(filePath string) (*rsa.PrivateKey, error) {
	keyData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

// NewApp constructs an App, loading the private key from the provided path.
func NewApp(appID, installationID, privateKeyPath string) (*App, error) {
	if appID == "" {
		return nil, fmt.Errorf("appID is required")
	}
	if installationID == "" {
		return nil, fmt.Errorf("installationID is required")
	}
	if privateKeyPath == "" {
		privateKeyPath = "/etc/secrets/github-app-private-key.pem"
	}

	priv, err := LoadPrivateKey(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	return &App{
		AppID:          appID,
		InstallationID: installationID,
		PrivateKey:     priv,
	}, nil
}

// GenerateJWT generates a JWT token for the GitHub App.
func (app *App) GenerateJWT() (string, error) {
	now := time.Now()
	claims := &jwt.StandardClaims{
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(10 * time.Minute).Unix(), // JWT expiration time
		Issuer:    app.AppID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(app.PrivateKey)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

// FetchInstallationToken retrieves the installation token and expiry for the given installation ID.
func (app *App) FetchInstallationToken(installationID string) (*TokenResponse, error) {
	jwtToken, err := app.GenerateJWT()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.github.com/app/installations/%s/access_tokens", installationID)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch installation token: %s", resp.Status)
	}

	var result struct {
		Token     string `json:"token"`
		ExpiresAt string `json:"expires_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// parse expiry
	var expiry time.Time
	if result.ExpiresAt != "" {
		expiry, _ = time.Parse(time.RFC3339, result.ExpiresAt)
	}

	return &TokenResponse{
		Token:     result.Token,
		ExpiresAt: expiry,
	}, nil
}

// GetInstallationToken uses the App's stored installation ID to fetch a fresh token.
func (app *App) GetInstallationToken() (*TokenResponse, error) {
	if app.InstallationID == "" {
		return nil, fmt.Errorf("installation ID not configured")
	}
	return app.FetchInstallationToken(app.InstallationID)
}