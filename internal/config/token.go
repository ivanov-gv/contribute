package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ivanov-gv/gh-contribute/internal/client/auth"
)

// ErrNotAuthenticated is returned when no token is found.
var ErrNotAuthenticated = errors.New(
	"not authenticated — run 'gh contribute auth login' first",
)

const (
	// TokenEnv is the environment variable checked first, suitable for CI / non-interactive use.
	TokenEnv = "GH_CONTRIBUTE_TOKEN"

	// tokenConfigPath is the token file path relative to the user's home directory.
	tokenConfigPath = ".config/gh-contribute/token"

	// configDirPermissions is the permission mode for the config directory (owner-only access).
	configDirPermissions = 0700

	// tokenFilePermissions is the permission mode for the token file (owner-only read/write).
	tokenFilePermissions = 0600
)

// LoadToken returns the GitHub access token.
// Priority: GH_CONTRIBUTE_TOKEN env var → GitHub App credentials → ~/.config/gh-contribute/token file.
func LoadToken() (string, error) {
	// 1. check env var first — CI / non-interactive environments
	if t := os.Getenv(TokenEnv); t != "" {
		return t, nil
	}

	// 2. try GitHub App auth (APP_ID + PRIVATE_KEY → installation token)
	token, err := tryAppAuth()
	if err != nil {
		return "", fmt.Errorf("tryAppAuth: %w", err)
	}
	if token != "" {
		return token, nil
	}

	// 3. fall back to config file (Device Flow token)
	return loadTokenFromFile()
}

// loadTokenFromFile reads the token from ~/.config/gh-contribute/token
func loadTokenFromFile() (string, error) {
	path, err := tokenFilePath()
	if err != nil {
		return "", fmt.Errorf("tokenFilePath: %w", err)
	}

	data, err := os.ReadFile(path)
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
// Returns ("", nil) if app auth is not configured (no APP_ID env var).
func tryAppAuth() (string, error) {
	appCfg, err := auth.LoadAppConfig()
	if err != nil {
		return "", fmt.Errorf("auth.LoadAppConfig: %w", err)
	}
	if appCfg == nil {
		return "", nil // not configured — skip
	}

	token, _, err := auth.GetAppToken(appCfg)
	if err != nil {
		return "", fmt.Errorf("auth.GetAppToken [appID=%d]: %w", appCfg.AppID, err)
	}

	return token, nil
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
