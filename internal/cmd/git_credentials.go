package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ivanov-gv/contribute/internal/config"
)

const githubHost = "github.com"

// newGitCredentialsCmd implements the git credential helper protocol for GitHub App tokens.
// Git calls this command as: contribute git-credentials <operation>
// where operation is one of: get, store, erase.
// Only "get" is handled; "store" and "erase" are no-ops since tokens are managed by the app.
func newGitCredentialsCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "git-credentials",
		Short:  "Git credential helper that provides GitHub App tokens (invoked by git automatically)",
		Hidden: true,
		Args:   cobra.MaximumNArgs(1),
		// skip app initialization — credential helper must work standalone without a git remote
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return nil },
		RunE: func(cmd *cobra.Command, args []string) error {
			// git calls with "get", "store", or "erase" — only handle "get"
			if len(args) == 0 || args[0] != "get" {
				return nil
			}
			return runCredentialGet()
		},
	}
}

// runCredentialGet reads a git credential request from stdin and writes the GitHub App
// installation token to stdout when the host is github.com.
// Exits silently (no output, no error) if not authenticated or the host is not GitHub,
// so git falls through to other configured helpers or prompts the user interactively.
func runCredentialGet() error {
	// read credential request fields from stdin (protocol=https\nhost=github.com\n\n)
	fields := map[string]string{}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		// each line is key=value
		key, value, ok := strings.Cut(line, "=")
		if ok {
			fields[key] = value
		}
	}

	// only serve credentials for github.com
	if fields["host"] != githubHost {
		return nil
	}

	// load the active token — silently exit if not configured so git can try other helpers
	token, err := config.LoadToken()
	if err != nil {
		if errors.Is(err, config.ErrNotAuthenticated) {
			return nil
		}
		return fmt.Errorf("config.LoadToken: %w", err)
	}

	// write credentials in git credential protocol format
	fmt.Fprintf(os.Stdout, "username=x-access-token\npassword=%s\n", token) //nolint:errcheck // writing to stdout
	return nil
}
