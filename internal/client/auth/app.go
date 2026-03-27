package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// jwtExpiry is the maximum lifetime for a GitHub App JWT (10 minutes per GitHub docs)
	jwtExpiry = 10 * time.Minute

	// jwtClockSkew accounts for clock drift between the host and GitHub servers
	jwtClockSkew = 60 * time.Second

	// githubInstallationsURL is the endpoint to list app installations
	githubInstallationsURL = "https://api.github.com/app/installations"
)

// AppConfig holds GitHub App credentials
type AppConfig struct {
	AppID          int64
	PrivateKey     *rsa.PrivateKey
	InstallationID int64 // 0 means auto-detect
}

// LoadAppConfig creates AppConfig from environment variables.
// Required: GH_CONTRIBUTE_APP_ID and either GH_CONTRIBUTE_PRIVATE_KEY (base64-encoded PEM)
// or GH_CONTRIBUTE_PRIVATE_KEY_PATH (file path to PEM).
// Optional: GH_CONTRIBUTE_INSTALLATION_ID (auto-detected if not set).
func LoadAppConfig() (*AppConfig, error) {
	appIDStr := os.Getenv("GH_CONTRIBUTE_APP_ID")
	if appIDStr == "" {
		return nil, nil // no app auth configured — fall through to token auth
	}

	var appID int64
	if _, err := fmt.Sscanf(appIDStr, "%d", &appID); err != nil {
		return nil, fmt.Errorf("invalid GH_CONTRIBUTE_APP_ID '%s': %w", appIDStr, err)
	}

	// load private key
	key, err := loadPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("loadPrivateKey: %w", err)
	}

	// optional installation ID
	var installationID int64
	if idStr := os.Getenv("GH_CONTRIBUTE_INSTALLATION_ID"); idStr != "" {
		if _, err := fmt.Sscanf(idStr, "%d", &installationID); err != nil {
			return nil, fmt.Errorf("invalid GH_CONTRIBUTE_INSTALLATION_ID '%s': %w", idStr, err)
		}
	}

	return &AppConfig{
		AppID:          appID,
		PrivateKey:     key,
		InstallationID: installationID,
	}, nil
}

// loadPrivateKey reads the RSA private key from env var or file
func loadPrivateKey() (*rsa.PrivateKey, error) {
	// try base64-encoded PEM from env var
	if encoded := os.Getenv("GH_CONTRIBUTE_PRIVATE_KEY"); encoded != "" {
		pemBytes, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			return nil, fmt.Errorf("base64.DecodeString GH_CONTRIBUTE_PRIVATE_KEY: %w", err)
		}
		return parsePrivateKey(pemBytes)
	}

	// try file path
	if path := os.Getenv("GH_CONTRIBUTE_PRIVATE_KEY_PATH"); path != "" {
		pemBytes, err := os.ReadFile(path) //nolint:gosec // path from trusted env var GH_CONTRIBUTE_PRIVATE_KEY_PATH
		if err != nil {
			return nil, fmt.Errorf("os.ReadFile [path='%s']: %w", path, err)
		}
		return parsePrivateKey(pemBytes)
	}

	return nil, fmt.Errorf("either GH_CONTRIBUTE_PRIVATE_KEY or GH_CONTRIBUTE_PRIVATE_KEY_PATH must be set")
}

// parsePrivateKey parses PEM-encoded RSA private key bytes
func parsePrivateKey(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("no PEM block found in private key")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// try PKCS8 format
		parsed, pkcs8Err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if pkcs8Err != nil {
			return nil, fmt.Errorf("failed to parse private key (PKCS1: %w, PKCS8: %w)", err, pkcs8Err)
		}
		rsaKey, ok := parsed.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("PKCS8 key is not RSA")
		}
		return rsaKey, nil
	}
	return key, nil
}

// GenerateJWT creates a signed JWT for GitHub App authentication
func GenerateJWT(appID int64, key *rsa.PrivateKey) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now.Add(-jwtClockSkew)),
		ExpiresAt: jwt.NewNumericDate(now.Add(jwtExpiry)),
		Issuer:    fmt.Sprintf("%d", appID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(key)
	if err != nil {
		return "", fmt.Errorf("token.SignedString: %w", err)
	}
	return signed, nil
}

// installationTokenResponse holds the response from the installation token endpoint
type installationTokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// GetInstallationToken exchanges a JWT for an installation access token.
// If installationID is 0, auto-detects the first installation.
func GetInstallationToken(jwtToken string, installationID int64) (string, time.Time, error) {
	if installationID == 0 {
		id, err := findInstallation(jwtToken)
		if err != nil {
			return "", time.Time{}, fmt.Errorf("findInstallation: %w", err)
		}
		installationID = id
	}

	url := fmt.Sprintf("https://api.github.com/app/installations/%d/access_tokens", installationID)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, nil)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("http.NewRequest: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("http.DefaultClient.Do: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // best-effort close on HTTP response body

	if resp.StatusCode != http.StatusCreated {
		return "", time.Time{}, fmt.Errorf("installation token request returned status %d", resp.StatusCode)
	}

	var result installationTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", time.Time{}, fmt.Errorf("json.Decode: %w", err)
	}

	return result.Token, result.ExpiresAt, nil
}

// installationNode holds minimal installation data for auto-detection
type installationNode struct {
	ID int64 `json:"id"`
}

// findInstallation returns the first installation ID for the app
func findInstallation(jwtToken string) (int64, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, githubInstallationsURL, nil)
	if err != nil {
		return 0, fmt.Errorf("http.NewRequest: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+jwtToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("http.DefaultClient.Do: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck // best-effort close on HTTP response body

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("installations request returned status %d", resp.StatusCode)
	}

	var installations []installationNode
	if err := json.NewDecoder(resp.Body).Decode(&installations); err != nil {
		return 0, fmt.Errorf("json.Decode: %w", err)
	}
	if len(installations) == 0 {
		return 0, fmt.Errorf("no installations found for this app")
	}

	return installations[0].ID, nil
}

// GetAppToken generates a JWT and exchanges it for an installation token.
// This is the high-level function that combines GenerateJWT + GetInstallationToken.
func GetAppToken(cfg *AppConfig) (string, time.Time, error) {
	jwtToken, err := GenerateJWT(cfg.AppID, cfg.PrivateKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("GenerateJWT: %w", err)
	}

	token, expiresAt, err := GetInstallationToken(jwtToken, cfg.InstallationID)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("GetInstallationToken: %w", err)
	}

	return token, expiresAt, nil
}
