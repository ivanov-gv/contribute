package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRemoteURL(t *testing.T) {
	tests := []struct {
		name          string
		remote        string
		expectedOwner string
		expectedRepo  string
		expectErr     bool
	}{
		{
			name:          "SSH standard",
			remote:        "git@github.com:ivanov-gv/gh-contribute.git",
			expectedOwner: "ivanov-gv",
			expectedRepo:  "gh-contribute",
		},
		{
			name:          "SSH without .git suffix",
			remote:        "git@github.com:owner/repo",
			expectedOwner: "owner",
			expectedRepo:  "repo",
		},
		{
			name:          "HTTPS standard",
			remote:        "https://github.com/ivanov-gv/gh-contribute.git",
			expectedOwner: "ivanov-gv",
			expectedRepo:  "gh-contribute",
		},
		{
			name:          "HTTPS without .git suffix",
			remote:        "https://github.com/owner/repo",
			expectedOwner: "owner",
			expectedRepo:  "repo",
		},
		{
			name:          "HTTP remote",
			remote:        "http://github.com/owner/repo.git",
			expectedOwner: "owner",
			expectedRepo:  "repo",
		},
		{
			name:          "proxy path SSH",
			remote:        "git@proxy.example.com:git/owner/repo.git",
			expectedOwner: "owner",
			expectedRepo:  "repo",
		},
		{
			name:          "proxy path HTTPS",
			remote:        "https://proxy.example.com/git/owner/repo.git",
			expectedOwner: "owner",
			expectedRepo:  "repo",
		},
		{
			name:      "invalid SSH format",
			remote:    "git@github.com",
			expectErr: true,
		},
		{
			name:      "invalid HTTPS — no path",
			remote:    "https://github.com",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := parseRemoteURL(tt.remote)
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedOwner, owner)
			assert.Equal(t, tt.expectedRepo, repo)
		})
	}
}

func TestParseOwnerRepo(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		expectedOwner string
		expectedRepo  string
		expectErr     bool
	}{
		{"standard", "owner/repo.git", "owner", "repo", false},
		{"no .git", "owner/repo", "owner", "repo", false},
		{"deep path", "/git/owner/repo", "owner", "repo", false},
		{"leading slash", "/owner/repo.git", "owner", "repo", false},
		{"single segment", "repo", "", "", true},
		{"empty", "", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := parseOwnerRepo(tt.path)
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedOwner, owner)
			assert.Equal(t, tt.expectedRepo, repo)
		})
	}
}
