package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadToken_EnvVar(t *testing.T) {
	t.Setenv(TokenEnv, "test-token-from-env")
	token, err := LoadToken()
	require.NoError(t, err)
	assert.Equal(t, "test-token-from-env", token)
}

func TestLoadToken_EnvVarTakesPriority(t *testing.T) {
	// even if a token file exists, env var wins
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "token")
	require.NoError(t, os.WriteFile(tokenPath, []byte("file-token"), 0600))

	t.Setenv(TokenEnv, "env-token")
	token, err := LoadToken()
	require.NoError(t, err)
	assert.Equal(t, "env-token", token)
}

func TestLoadToken_MissingFile(t *testing.T) {
	t.Setenv(TokenEnv, "")
	// override HOME to a temp dir with no token file
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	_, err := LoadToken()
	assert.ErrorIs(t, err, ErrNotAuthenticated)
}

func TestLoadToken_EmptyFile(t *testing.T) {
	t.Setenv(TokenEnv, "")
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// create empty token file
	configDir := filepath.Join(tmpDir, ".config", "gh-contribute")
	require.NoError(t, os.MkdirAll(configDir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "token"), []byte(""), 0600))

	_, err := LoadToken()
	assert.ErrorIs(t, err, ErrNotAuthenticated)
}

func TestLoadToken_WhitespaceFile(t *testing.T) {
	t.Setenv(TokenEnv, "")
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// create token file with whitespace
	configDir := filepath.Join(tmpDir, ".config", "gh-contribute")
	require.NoError(t, os.MkdirAll(configDir, 0700))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "token"), []byte("  \n  "), 0600))

	_, err := LoadToken()
	assert.ErrorIs(t, err, ErrNotAuthenticated)
}

func TestSaveAndLoadToken(t *testing.T) {
	t.Setenv(TokenEnv, "")
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// save
	require.NoError(t, SaveToken("my-secret-token"))

	// load
	token, err := LoadToken()
	require.NoError(t, err)
	assert.Equal(t, "my-secret-token", token)

	// verify file permissions
	path := filepath.Join(tmpDir, ".config", "gh-contribute", "token")
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}
