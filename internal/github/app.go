package github

import (
    "crypto/rsa"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "time"

    "github.com/dgrijalva/jwt-go"
)

type App struct {
    AppID      string
    PrivateKey *rsa.PrivateKey
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
    claims := &jwt.StandardClaims{
        IssuedAt:  now.Unix(),
        ExpiresAt: now.Add(10 * time.Minute).Unix(), // JWT expiration time
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