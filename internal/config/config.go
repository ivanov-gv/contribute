package config

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"

	"github.com/ivanov-gv/contribute/internal/client/auth"
)

// Config holds the runtime configuration.
type Config struct {
	Token    string              // initial GitHub token
	Provider *auth.TokenProvider // non-nil when using GitHub App auth; handles automatic token refresh
	Owner    string              // repository owner
	Repo     string              // repository name
}

// Load reads configuration from the environment and git context.
// Token priority: GH_CONTRIBUTE_TOKEN env var → GitHub App credentials → ~/.config/contribute/token file.
func Load() (*Config, error) {
	// load .env if present (ignore error — file is optional)
	_ = godotenv.Load()

	// load token; for App auth, also get the TokenProvider for automatic refresh
	provider, token, err := loadTokenWithProvider()
	if err != nil {
		return nil, fmt.Errorf("loadTokenWithProvider: %w", err)
	}

	// detect owner/repo from git remote
	owner, repo, err := detectRepo()
	if err != nil {
		return nil, fmt.Errorf("detectRepo: %w", err)
	}

	return &Config{
		Token:    token,
		Provider: provider,
		Owner:    owner,
		Repo:     repo,
	}, nil
}

// detectRepo extracts owner/repo from the git remote "origin".
func detectRepo() (string, string, error) {
	out, err := exec.CommandContext(context.Background(), "git", "remote", "get-url", "origin").Output()
	if err != nil {
		return "", "", fmt.Errorf("git remote get-url origin: %w", err)
	}

	remote := strings.TrimSpace(string(out))
	return parseRemoteURL(remote)
}

// parseRemoteURL extracts owner/repo from SSH or HTTPS remote URLs.
func parseRemoteURL(remote string) (string, string, error) {
	// SSH: git@github.com:owner/repo.git
	if strings.HasPrefix(remote, "git@") {
		parts := strings.SplitN(remote, ":", 2) //nolint:mnd // split into host:path
		if len(parts) != 2 {                    //nolint:mnd // expect host and path
			return "", "", fmt.Errorf("unexpected SSH remote format: %s", remote)
		}
		return parseOwnerRepo(parts[1])
	}

	// HTTPS: https://github.com/owner/repo.git
	remote = strings.TrimPrefix(remote, "https://")
	remote = strings.TrimPrefix(remote, "http://")
	// remove host part
	parts := strings.SplitN(remote, "/", 2) //nolint:mnd // split into host/path
	if len(parts) != 2 {                    //nolint:mnd // expect host and path
		return "", "", fmt.Errorf("unexpected HTTPS remote format: %s", remote)
	}
	return parseOwnerRepo(parts[1])
}

// parseOwnerRepo extracts "owner/repo" from the last two slash-separated path segments.
// Handles standard paths ("owner/repo.git") and proxy paths ("/git/owner/repo").
func parseOwnerRepo(path string) (string, string, error) {
	path = strings.TrimSuffix(path, ".git")
	parts := strings.Split(path, "/")
	// filter out empty segments (leading slash, etc.)
	var segments []string
	for _, p := range parts {
		if p != "" {
			segments = append(segments, p)
		}
	}
	if len(segments) < 2 { //nolint:mnd // need at least owner and repo
		return "", "", fmt.Errorf("cannot parse owner/repo from: %s", path)
	}
	owner := segments[len(segments)-2]
	repo := segments[len(segments)-1]
	return owner, repo, nil
}
