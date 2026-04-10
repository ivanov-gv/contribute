package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ivanov-gv/gh-contribute/internal/config"
)

// newTokenCmd prints a valid GitHub token to stdout for use with other tools.
// Intended for composing with gh CLI: GH_TOKEN=$(gh contribute token) gh pr view 123
func newTokenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "token",
		Short: "Print a valid GitHub token to stdout",
		Long: `Print the active GitHub token to stdout for use with other tools.

The token follows the same priority chain as all gh-contribute commands:
GH_CONTRIBUTE_TOKEN env var → GitHub App credentials (auto-refreshed).

Example:
  GH_TOKEN=$(gh contribute token) gh pr view 123
  GH_TOKEN=$(gh contribute token) gh api /user`,
		// skip app initialization — token command does not need owner/repo detection
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return nil },
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := config.LoadToken()
			if err != nil {
				return fmt.Errorf("token: %w", err)
			}

			fmt.Fprint(os.Stdout, token) //nolint:errcheck // writing to stdout
			return nil
		},
	}
}
