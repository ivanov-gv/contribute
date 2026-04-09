package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	// appConfigPath is the app credentials file path relative to the user's home directory.
	appConfigPath = ".config/gh-contribute/app.json"

	// appConfigFilePermissions is the permission mode for the app config file (owner-only read/write).
	appConfigFilePermissions = 0600
)

// storedAppConfig holds GitHub App credentials persisted to disk.
type storedAppConfig struct {
	AppID          int64  `json:"app_id"`
	PrivateKeyPath string `json:"private_key_path"`
	InstallationID int64  `json:"installation_id,omitempty"`
}

// SaveAppCredentials persists the GitHub App credentials to ~/.config/gh-contribute/app.json.
func SaveAppCredentials(appID int64, keyPath string, installationID int64) error {
	path, err := appConfigFilePath()
	if err != nil {
		return fmt.Errorf("appConfigFilePath: %w", err)
	}

	// create parent directories with restricted permissions
	if err := os.MkdirAll(filepath.Dir(path), configDirPermissions); err != nil {
		return fmt.Errorf("os.MkdirAll [dir='%s']: %w", filepath.Dir(path), err)
	}

	data, err := json.MarshalIndent(storedAppConfig{
		AppID:          appID,
		PrivateKeyPath: keyPath,
		InstallationID: installationID,
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("json.MarshalIndent: %w", err)
	}

	if err := os.WriteFile(path, data, appConfigFilePermissions); err != nil {
		return fmt.Errorf("os.WriteFile [path='%s']: %w", path, err)
	}

	return nil
}

// LoadStoredAppCredentials reads the stored app credentials from ~/.config/gh-contribute/app.json
// and returns (AppID, PrivateKeyPath). Returns (0, "") when no file exists.
func LoadStoredAppCredentials() (appID int64, keyPath string, err error) {
	cfg, err := loadAppCredentials()
	if err != nil {
		return 0, "", fmt.Errorf("loadAppCredentials: %w", err)
	}
	if cfg == nil {
		return 0, "", nil
	}
	return cfg.AppID, cfg.PrivateKeyPath, nil
}

// loadAppCredentials reads the stored app credentials from ~/.config/gh-contribute/app.json.
// Returns nil, nil when the file does not exist (app auth not configured via CLI).
func loadAppCredentials() (*storedAppConfig, error) {
	path, err := appConfigFilePath()
	if err != nil {
		return nil, fmt.Errorf("appConfigFilePath: %w", err)
	}

	data, err := os.ReadFile(path) //nolint:gosec // path is from appConfigFilePath() which uses a constant relative to $HOME
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil // not configured
		}
		return nil, fmt.Errorf("os.ReadFile [path='%s']: %w", path, err)
	}

	var cfg storedAppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return &cfg, nil
}

// appConfigFilePath returns the absolute path to the app credentials config file.
func appConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("os.UserHomeDir: %w", err)
	}
	return filepath.Join(home, appConfigPath), nil
}
