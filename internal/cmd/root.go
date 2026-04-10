package cmd

import (
	"fmt"
	"os"

	ghrest "github.com/google/go-github/v69/github"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shurcooL/githubv4"
	"github.com/spf13/cobra"

	ghclient "github.com/ivanov-gv/contribute/internal/client/github"
	"github.com/ivanov-gv/contribute/internal/config"
	"github.com/ivanov-gv/contribute/internal/service/comment"
	"github.com/ivanov-gv/contribute/internal/service/issue"
	"github.com/ivanov-gv/contribute/internal/service/pr"
	"github.com/ivanov-gv/contribute/internal/service/reaction"
	"github.com/ivanov-gv/contribute/internal/service/review"
	"github.com/ivanov-gv/contribute/internal/service/thread"
)

// app holds shared dependencies for all authenticated commands.
type app struct {
	cfg             *config.Config
	prService       *pr.Service
	commentService  *comment.Service
	reactionService *reaction.Service
	reviewService   *review.Service
	threadService   *thread.Service
	issueService    *issue.Service
}

// init loads config and initializes all services.
// Called by the root PersistentPreRunE before any authenticated command runs.
func (a *app) init() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config.Load: %w", err)
	}

	var gql *githubv4.Client
	if cfg.Provider != nil {
		gql = ghclient.NewGraphQLClientWithProvider(cfg.Provider)
	} else {
		gql = ghclient.NewGraphQLClient(cfg.Token)
	}
	rest := ghrest.NewClient(nil).WithAuthToken(cfg.Token)

	log.Debug().Str("owner", cfg.Owner).Str("repo", cfg.Repo).Msg("config loaded")

	a.cfg = cfg
	a.prService = pr.NewService(gql, cfg.Owner, cfg.Repo)
	a.commentService = comment.NewService(gql, rest, cfg.Owner, cfg.Repo)
	a.reactionService = reaction.NewService(rest, cfg.Owner, cfg.Repo)
	a.reviewService = review.NewService(gql, cfg.Owner, cfg.Repo)
	a.threadService = thread.NewService(gql, cfg.Owner, cfg.Repo)
	a.issueService = issue.NewService(gql, cfg.Owner, cfg.Repo)

	return nil
}

// Execute wires and runs the root command.
func Execute() {
	// human-readable console output
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	_app := &app{}

	rootCmd := &cobra.Command{
		Use:          "contribute",
		Short:        "A CLI tool for simplifying agents interaction with PRs on GitHub",
		SilenceUsage: true,
		// initialize app before any authenticated command runs
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			verbose, _ := cmd.Flags().GetBool("verbose")
			if verbose {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			} else {
				zerolog.SetGlobalLevel(zerolog.InfoLevel)
			}
			return _app.init()
		},
	}

	rootCmd.PersistentFlags().String("format", "", "Output format: json (default: markdown)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable debug logging")

	rootCmd.AddCommand(
		// login overrides PersistentPreRunE with a no-op — no token required
		newLoginCmd(),
		// auth subcommand (status) overrides PersistentPreRunE with a no-op
		newAuthCmd(),
		// git-credentials is a hidden command implementing the git credential helper protocol
		newGitCredentialsCmd(),
		// token prints a valid token to stdout for use with gh CLI and other tools
		newTokenCmd(),
		// authenticated commands — app is initialized via PersistentPreRunE
		_app.newPRCmd(),
		_app.newCommentsCmd(),
		_app.newCommentCmd(),
		_app.newReactCmd(),
		_app.newReplyCmd(),
		_app.newResolveCmd(),
		_app.newUnresolveCmd(),
		_app.newReviewCmd(),
		_app.newReviewCommentCmd(),
		_app.newSubmitReviewCmd(),
		_app.newThreadCmd(),
		_app.newIssueCmd(),
		_app.newIssueCommentCmd(),
		_app.newIssuesCmd(),
		_app.newWatchCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
