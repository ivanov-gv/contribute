package git

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// CurrentBranch returns the name of the currently checked-out branch.
func CurrentBranch() (string, error) {
	out, err := exec.CommandContext(context.Background(), "git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --abbrev-ref HEAD: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// SetupCredentialHelper configures git (globally) to use the given shell command as a
// credential helper for GitHub HTTPS authentication. Scoped to github.com only so it
// does not interfere with other credential helpers on the system.
func SetupCredentialHelper(helperCmd string) error {
	cmd := exec.CommandContext(context.Background(), "git", "config", "--global", "--replace-all", //nolint:gosec // helperCmd is a hardcoded string from the call site, not user input
		"credential.https://github.com.helper", "!"+helperCmd)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git config --global credential.https://github.com.helper: %w: %s",
			err, strings.TrimSpace(string(out)))
	}
	return nil
}
