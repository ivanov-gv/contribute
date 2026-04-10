package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadTokenWithProvider_EnvVar(t *testing.T) {
	t.Setenv(TokenEnv, "test-token-from-env")
	_, token, err := loadTokenWithProvider()
	require.NoError(t, err)
	assert.Equal(t, "test-token-from-env", token)
}

func TestLoadTokenWithProvider_NotConfigured(t *testing.T) {
	// clear all auth env vars and point HOME to an empty temp dir (no app.json)
	t.Setenv(TokenEnv, "")
	t.Setenv("GH_CONTRIBUTE_APP_ID", "")
	t.Setenv("GH_CONTRIBUTE_PRIVATE_KEY", "")
	t.Setenv("GH_CONTRIBUTE_PRIVATE_KEY_PATH", "")
	t.Setenv("HOME", t.TempDir())

	_, _, err := loadTokenWithProvider()
	assert.ErrorIs(t, err, ErrNotAuthenticated)
}
