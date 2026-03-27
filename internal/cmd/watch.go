package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

const defaultWatchInterval = 30 * time.Second

func (a *app) newWatchCmd() *cobra.Command {
	var prNumber int
	var intervalStr string

	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Poll for new activity on a PR and print changes",
		Long:  "Polls the comments endpoint at a regular interval, printing new comments and reviews as they appear.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// resolve PR number
			number, err := a.resolvePR(prNumber)
			if err != nil {
				return fmt.Errorf("resolvePR [pr=%d]: %w", prNumber, err)
			}

			interval, err := time.ParseDuration(intervalStr)
			if err != nil {
				return fmt.Errorf("invalid interval '%s': %w", intervalStr, err)
			}

			return a.watchLoop(number, interval)
		},
	}

	cmd.Flags().IntVar(&prNumber, "pr", 0, "PR number (auto-detected from current branch if not set)")
	cmd.Flags().StringVar(&intervalStr, "interval", "30s", "Poll interval (e.g. 30s, 1m)")

	return cmd
}

// watchLoop polls for new comments and reviews, printing new items to stdout
func (a *app) watchLoop(prNumber int, interval time.Duration) error {
	// get initial state
	known, err := a.commentService.List(prNumber)
	if err != nil {
		return fmt.Errorf("commentService.List [pr=%d]: %w", prNumber, err)
	}

	// build set of known comment/review IDs
	knownIDs := make(map[int64]bool)
	for _, c := range known.IssueComments {
		knownIDs[c.DatabaseID] = true
	}
	for _, r := range known.Reviews {
		knownIDs[r.DatabaseID] = true
	}

	fmt.Printf("watching PR #%d (interval %s, %d existing items)\n", prNumber, interval, len(knownIDs))

	for {
		time.Sleep(interval)

		current, err := a.commentService.List(prNumber)
		if err != nil {
			fmt.Printf("error polling: %v\n", err)
			continue
		}

		// detect new issue comments
		for _, c := range current.IssueComments {
			if !knownIDs[c.DatabaseID] {
				knownIDs[c.DatabaseID] = true
				fmt.Printf("\n--- new comment #%d by @%s at %s ---\n%s\n", c.DatabaseID, c.Author, c.CreatedAt, c.Body)
			}
		}

		// detect new reviews
		for _, r := range current.Reviews {
			if !knownIDs[r.DatabaseID] {
				knownIDs[r.DatabaseID] = true
				fmt.Printf("\n--- new review #%d (%s) by @%s at %s ---\n", r.DatabaseID, r.State, r.Author, r.CreatedAt)
				if r.Body != "" {
					fmt.Println(r.Body)
				}
				if r.CommentCount > 0 {
					fmt.Printf("(%d inline comments — run: gh-contribute review %d --pr %d)\n", r.CommentCount, r.DatabaseID, prNumber)
				}
			}
		}
	}
}
