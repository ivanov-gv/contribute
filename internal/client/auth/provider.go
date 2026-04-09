package auth

import (
	"fmt"
	"sync"
	"time"
)

const tokenRefreshBuffer = 5 * time.Minute

// TokenProvider manages token lifecycle with automatic refresh.
// Safe for concurrent use.
type TokenProvider struct {
	appCfg    *AppConfig
	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

// NewTokenProvider creates a provider that generates and caches installation tokens.
func NewTokenProvider(appCfg *AppConfig) *TokenProvider {
	return &TokenProvider{appCfg: appCfg}
}

// Token returns a valid installation token, refreshing if needed.
func (p *TokenProvider) Token() (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// return cached token if still valid (with buffer)
	if p.token != "" && time.Now().Add(tokenRefreshBuffer).Before(p.expiresAt) {
		return p.token, nil
	}

	// refresh
	token, expiresAt, err := GetAppToken(p.appCfg)
	if err != nil {
		return "", fmt.Errorf("GetAppToken: %w", err)
	}

	p.token = token
	p.expiresAt = expiresAt
	return p.token, nil
}
