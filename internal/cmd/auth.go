package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/ivanov-gv/gh-contribute/internal/client/auth"
	"github.com/ivanov-gv/gh-contribute/internal/config"
)

// newAuthCmd returns the "auth" parent command with login-app and status subcommands.
func newAuthCmd() *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage GitHub App authentication",
		// skip app initialization — auth commands do not require a stored token
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return nil },
	}
	authCmd.AddCommand(newAuthLoginAppCmd(), newAuthStatusCmd())
	return authCmd
}

// newAuthLoginAppCmd stores GitHub App credentials for non-interactive authentication.
// The app must be installed on the target repository before tokens can be issued.
func newAuthLoginAppCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login-app",
		Short: "Authenticate as a GitHub App using a private key",
		Long: `Store GitHub App credentials for automatic installation-token authentication.

The App must be installed on your target repository. If --installation-id is
omitted, gh-contribute auto-detects the first installation.

--app-id can be omitted if GH_CONTRIBUTE_APP_ID is set in the environment.

Example:
  gh contribute auth login-app --app-id 123456 --key-path ~/.config/gh-contribute/private-key.pem
  GH_CONTRIBUTE_APP_ID=123456 gh contribute auth login-app --key-path ~/.config/gh-contribute/private-key.pem`,
		RunE: func(cmd *cobra.Command, args []string) error {
			const logfmt = "auth login-app: "

			appID, _ := cmd.Flags().GetInt64("app-id")
			keyPath, _ := cmd.Flags().GetString("key-path")
			installationID, _ := cmd.Flags().GetInt64("installation-id")

			// fall back to GH_CONTRIBUTE_APP_ID env var if flag not set
			if appID == 0 {
				if envVal := os.Getenv("GH_CONTRIBUTE_APP_ID"); envVal != "" {
					if _, err := fmt.Sscanf(envVal, "%d", &appID); err != nil {
						return fmt.Errorf(logfmt+"invalid GH_CONTRIBUTE_APP_ID '%s': %w", envVal, err)
					}
				}
			}
			if appID == 0 {
				return fmt.Errorf(logfmt + "--app-id is required (or set GH_CONTRIBUTE_APP_ID)")
			}
			if keyPath == "" {
				return fmt.Errorf(logfmt + "--key-path is required")
			}

			// expand ~ to home directory
			if len(keyPath) > 1 && keyPath[0] == '~' {
				home, err := os.UserHomeDir()
				if err != nil {
					return fmt.Errorf(logfmt+"os.UserHomeDir: %w", err)
				}
				keyPath = filepath.Join(home, keyPath[1:])
			}

			// validate credentials by obtaining an installation token before saving
			appCfg, err := auth.LoadAppConfigFromPath(appID, keyPath, installationID)
			if err != nil {
				return fmt.Errorf(logfmt+"auth.LoadAppConfigFromPath: %w", err)
			}

			appName, err := auth.GetAppName(context.Background(), appCfg.AppID, appCfg.PrivateKey)
			if err != nil {
				return fmt.Errorf(logfmt+"auth.GetAppName: %w", err)
			}

			_, _, err = auth.GetAppToken(appCfg)
			if err != nil {
				return fmt.Errorf(logfmt+"auth.GetAppToken: %w", err)
			}

			if err := config.SaveAppCredentials(appID, keyPath, installationID); err != nil {
				return fmt.Errorf(logfmt+"config.SaveAppCredentials: %w", err)
			}

			log.Info().
				Str("app", appName).
				Int64("app_id", appID).
				Msg(logfmt + "authenticated successfully")
			return nil
		},
	}

	cmd.Flags().Int64("app-id", 0, "GitHub App ID (required)")
	cmd.Flags().String("key-path", "", "Path to the GitHub App private key PEM file (required)")
	cmd.Flags().Int64("installation-id", 0, "Installation ID (optional, auto-detected if not set)")

	return cmd
}

// newAuthStatusCmd prints the app name and ID for the active GitHub App authentication.
func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			const logfmt = "auth status: "

			// load app config from env vars (highest priority) or stored file
			appCfg, err := config.LoadAppConfig()
			if err != nil {
				return fmt.Errorf(logfmt+"config.LoadAppConfig: %w", err)
			}
			if appCfg == nil {
				return fmt.Errorf(logfmt + "not authenticated — set GH_CONTRIBUTE_APP_ID and GH_CONTRIBUTE_PRIVATE_KEY_PATH, or run 'gh contribute auth login-app'")
			}

			appName, err := auth.GetAppName(context.Background(), appCfg.AppID, appCfg.PrivateKey)
			if err != nil {
				return fmt.Errorf(logfmt+"auth.GetAppName: %w", err)
			}

			log.Info().Str("app", appName).Int64("app_id", appCfg.AppID).Msg(logfmt + "logged in as app")
			return nil
		},
	}
}
