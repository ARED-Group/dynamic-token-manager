package github

import (
    "crypto/rsa"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

type App struct {
    AppID          string
    InstallationID string
    PrivateKey     *rsa.PrivateKey
}

// NewApp creates a new GitHub App instance
func NewApp(appID, installationID, privateKeyPath string) (*App, error) {
    privateKey, err := LoadPrivateKey(privateKeyPath)
    if err != nil {
        return nil, fmt.Errorf("failed to load private key: %w", err)
    }
    
    return &App{
        AppID:          appID,
        InstallationID: installationID,
        PrivateKey:     privateKey,
    }, nil
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

// GenerateJWT generates a JWT token for the GitHub App.
func (app *App) GenerateJWT() (string, error) {
    now := time.Now()
    claims := &jwt.RegisteredClaims{
        IssuedAt:  jwt.NewNumericDate(now),
        ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)), // JWT expiration time
        Subject:   app.AppID,
    }

    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    signedToken, err := token.SignedString(app.PrivateKey)
    if err != nil {
        return "", err
    }
    return signedToken, nil
}

// FetchInstallationToken retrieves the installation token for the GitHub App.
func (app *App) FetchInstallationToken(installationID string) (string, error) {
    jwtToken, err := app.GenerateJWT()
    if err != nil {
        return "", err
    }

    url := fmt.Sprintf("https://api.github.com/app/installations/%s/access_tokens", installationID)
    req, err := http.NewRequest("POST", url, nil)
    if err != nil {
        return "", err
    }
    req.Header.Set("Authorization", "Bearer "+jwtToken)
    req.Header.Set("Accept", "application/vnd.github.v3+json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusCreated {
        return "", fmt.Errorf("failed to fetch installation token: %s", resp.Status)
    }

    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", err
    }

    token, ok := result["token"].(string)
    if !ok {
        return "", fmt.Errorf("token not found in response")
    }
    return token, nil
}

// GitHubTokenResponse represents a GitHub token response
type GitHubTokenResponse struct {
    Token     string    `json:"token"`
    ExpiresAt time.Time `json:"expires_at"`
}

// GetInstallationToken gets the installation token with structured response
func (app *App) GetInstallationToken() (*GitHubTokenResponse, error) {
    token, err := app.FetchInstallationToken(app.InstallationID)
    if err != nil {
        return nil, err
    }
    
    // GitHub tokens typically expire in 1 hour
    return &GitHubTokenResponse{
        Token:     token,
        ExpiresAt: time.Now().Add(time.Hour),
    }, nil
}