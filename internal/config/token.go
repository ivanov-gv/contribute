package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ivanov-gv/gh-contribute/internal/client/auth"
)

const (
	// TokenEnv is the environment variable checked first, suitable for CI / non-interactive use.
	TokenEnv = "GH_CONTRIBUTE_TOKEN"

	// tokenConfigPath is the token file path relative to the user's home directory.
	tokenConfigPath = ".config/gh-contribute/token" //nolint:gosec // not a credential, it's the path where the token is stored

	// configDirPermissions is the permission mode for the config directory (owner-only access).
	configDirPermissions = 0700

	// tokenFilePermissions is the permission mode for the token file (owner-only read/write).
	tokenFilePermissions = 0600
)

// LoadToken returns the GitHub access token.
// Priority: GH_CONTRIBUTE_TOKEN env var → GitHub App credentials → ~/.config/gh-contribute/token file.
func LoadToken() (string, error) {
	_, token, err := loadTokenWithProvider()
	return token, err
}

// loadTokenWithProvider returns the token and, when using GitHub App auth, the TokenProvider
// that handles automatic token refresh. Provider is nil for env var and file-based tokens.
func loadTokenWithProvider() (*auth.TokenProvider, string, error) {
	// 1. check env var first — CI / non-interactive environments
	if t := os.Getenv(TokenEnv); t != "" {
		return nil, t, nil
	}

	// 2. try GitHub App auth (APP_ID + PRIVATE_KEY → installation token via TokenProvider)
	provider, token, err := tryAppAuth()
	if err != nil {
		return nil, "", fmt.Errorf("tryAppAuth: %w", err)
	}
	if token != "" {
		return provider, token, nil
	}

	// 3. fall back to config file (Device Flow token)
	token, err = loadTokenFromFile()
	return nil, token, err
}

// loadTokenFromFile reads the token from ~/.config/gh-contribute/token
func loadTokenFromFile() (string, error) {
	path, err := tokenFilePath()
	if err != nil {
		return "", fmt.Errorf("tokenFilePath: %w", err)
	}

	data, err := os.ReadFile(path) //nolint:gosec // path is from tokenFilePath() which uses a constant relative to $HOME
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrNotAuthenticated
		}
		return "", fmt.Errorf("os.ReadFile [path='%s']: %w", path, err)
	}

	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", ErrNotAuthenticated
	}

	return token, nil
}

// tryAppAuth attempts GitHub App authentication if credentials are configured.
// Priority: env vars (GH_CONTRIBUTE_APP_ID) → stored app.json credentials.
// Returns (nil, "", nil) if app auth is not configured.
// On success, returns a TokenProvider that caches and refreshes the token automatically,
// plus the initial token string for immediate use.
func tryAppAuth() (*auth.TokenProvider, string, error) {
	// 1. try env vars
	appCfg, err := auth.LoadAppConfig()
	if err != nil {
		return nil, "", fmt.Errorf("auth.LoadAppConfig: %w", err)
	}

	// 2. fall back to stored app credentials file
	if appCfg == nil {
		appCfg, err = loadStoredAppConfig()
		if err != nil {
			return nil, "", fmt.Errorf("loadStoredAppConfig: %w", err)
		}
	}

	if appCfg == nil {
		return nil, "", nil // not configured — skip
	}

	provider := auth.NewTokenProvider(appCfg)
	token, err := provider.Token()
	if err != nil {
		return nil, "", fmt.Errorf("provider.Token [appID=%d]: %w", appCfg.AppID, err)
	}

	return provider, token, nil
}

// loadStoredAppConfig reads app credentials from the config file and converts them to AppConfig.
// Returns nil, nil when no config file exists.
func loadStoredAppConfig() (*auth.AppConfig, error) {
	stored, err := loadAppCredentials()
	if err != nil {
		return nil, fmt.Errorf("loadAppCredentials: %w", err)
	}
	if stored == nil {
		return nil, nil // not configured
	}

	return auth.LoadAppConfigFromPath(stored.AppID, stored.PrivateKeyPath, stored.InstallationID)
}

// SaveToken persists the token to ~/.config/gh-contribute/token with 0600 permissions.
func SaveToken(token string) error {
	path, err := tokenFilePath()
	if err != nil {
		return fmt.Errorf("tokenFilePath: %w", err)
	}

	// create parent directories with restricted permissions
	if err := os.MkdirAll(filepath.Dir(path), configDirPermissions); err != nil {
		return fmt.Errorf("os.MkdirAll [dir='%s']: %w", filepath.Dir(path), err)
	}

	// write with owner-only permissions
	if err := os.WriteFile(path, []byte(token), tokenFilePermissions); err != nil {
		return fmt.Errorf("os.WriteFile [path='%s']: %w", path, err)
	}

	return nil
}

// tokenFilePath returns the absolute path to the token config file.
func tokenFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("os.UserHomeDir: %w", err)
	}
	return filepath.Join(home, tokenConfigPath), nil
}
