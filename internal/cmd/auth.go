package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/ivanov-gv/gh-contribute/internal/client/auth"
	"github.com/ivanov-gv/gh-contribute/internal/config"
)

// newAuthCmd returns the "auth" parent command with login, login-app, and status subcommands.
func newAuthCmd() *cobra.Command {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage GitHub App authentication",
		// skip app initialization — auth commands do not require a stored token
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return nil },
	}
	authCmd.AddCommand(newAuthLoginCmd(), newAuthLoginAppCmd(), newAuthStatusCmd())
	return authCmd
}

// newAuthLoginCmd initiates the Device Authorization Flow and stores the resulting token.
func newAuthLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate via GitHub App Device Authorization Flow",
		RunE: func(cmd *cobra.Command, args []string) error {
			const logfmt = "auth login: "

			token, err := auth.RunDeviceFlow(func(userCode, verificationURI string) {
				// print URL and code to stdout so the user can act on them
				fmt.Printf("Open: %s\nEnter code: %s\n", verificationURI, userCode)

				// best-effort: try to open the browser
				openBrowser(verificationURI)

				log.Info().Msg(logfmt + "waiting for authorization...")
			})
			if err != nil {
				return fmt.Errorf(logfmt+"auth.RunDeviceFlow: %w", err)
			}

			if err := config.SaveToken(token); err != nil {
				return fmt.Errorf(logfmt+"config.SaveToken: %w", err)
			}

			log.Info().Msg(logfmt + "authentication successful")
			return nil
		},
	}
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

// newAuthStatusCmd prints the authentication identity (username for user tokens, app name for app tokens).
func newAuthStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		RunE: func(cmd *cobra.Command, args []string) error {
			const logfmt = "auth status: "

			// check for stored app credentials first — app tokens cannot call /user
			appID, keyPath, err := config.LoadStoredAppCredentials()
			if err != nil {
				return fmt.Errorf(logfmt+"config.LoadStoredAppCredentials: %w", err)
			}
			if appID != 0 && keyPath != "" {
				appCfg, err := auth.LoadAppConfigFromPath(appID, keyPath, 0)
				if err != nil {
					return fmt.Errorf(logfmt+"auth.LoadAppConfigFromPath: %w", err)
				}
				name, err := auth.GetAppName(context.Background(), appCfg.AppID, appCfg.PrivateKey)
				if err != nil {
					return fmt.Errorf(logfmt+"auth.GetAppName: %w", err)
				}
				log.Info().Str("app", name).Int64("app_id", appID).Msg(logfmt + "logged in as app")
				return nil
			}

			// fall back to user token
			token, err := config.LoadToken()
			if err != nil {
				return fmt.Errorf(logfmt+"config.LoadToken: %w", err)
			}

			username, err := auth.GetUsername(context.Background(), token)
			if err != nil {
				return fmt.Errorf(logfmt+"auth.GetUsername: %w", err)
			}

			log.Info().Str("username", username).Msg(logfmt + "logged in")
			return nil
		},
	}
}

// openBrowser attempts to open uri in the user's default browser.
// Failures are logged at debug level — the user can always open the URL manually.
func openBrowser(uri string) {
	var cmd string
	switch runtime.GOOS {
	case "linux":
		cmd = "xdg-open"
	case "darwin":
		cmd = "open"
	default:
		log.Debug().Str("os", runtime.GOOS).Msg("openBrowser: unsupported OS, please open URL manually")
		return
	}
	if err := exec.CommandContext(context.Background(), cmd, uri).Start(); err != nil { //nolint:gosec // cmd is a trusted constant ("xdg-open" or "open")
		log.Debug().Err(err).Msg("openBrowser: could not open browser, please open URL manually")
	}
}
