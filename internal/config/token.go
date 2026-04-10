package config

import (
	"fmt"
	"os"

	"github.com/ivanov-gv/contribute/internal/client/auth"
)

const (
	// TokenEnv is the environment variable checked first, suitable for CI / non-interactive use.
	TokenEnv = "GH_CONTRIBUTE_TOKEN"

	// configDirPermissions is the permission mode for the config directory (owner-only access).
	configDirPermissions = 0700
)

// loadTokenWithProvider returns the token and, when using GitHub App auth, the TokenProvider
// that handles automatic token refresh. Provider is nil for direct token env var.
// Returns ErrNotAuthenticated when no credentials are configured.
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

	// 3. no credentials configured — prompt the user
	return nil, "", ErrNotAuthenticated
}

// LoadToken returns the active GitHub token using the priority chain.
// Returns ErrNotAuthenticated when no credentials are configured.
func LoadToken() (string, error) {
	provider, token, err := loadTokenWithProvider()
	if err != nil {
		return "", fmt.Errorf("loadTokenWithProvider: %w", err)
	}
	if provider != nil {
		// provider handles automatic refresh
		t, refreshErr := provider.Token()
		if refreshErr != nil {
			return "", fmt.Errorf("provider.Token: %w", refreshErr)
		}
		return t, nil
	}
	return token, nil
}

// tryAppAuth attempts GitHub App authentication if credentials are configured.
// Returns (nil, "", nil) if app auth is not configured.
// On success, returns a TokenProvider that caches and refreshes the token automatically,
// plus the initial token string for immediate use.
func tryAppAuth() (*auth.TokenProvider, string, error) {
	appCfg, err := LoadAppConfig()
	if err != nil {
		return nil, "", fmt.Errorf("LoadAppConfig: %w", err)
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
