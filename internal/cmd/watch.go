package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"
)

const (
	defaultWatchInterval  = 30 * time.Second
	watchMaxConsecErrors  = 5
	watchMaxKnownIDs      = 1000
	watchErrorBackoffBase = 5 * time.Second
)

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

			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
			defer stop()

			return a.watchLoop(ctx, number, interval)
		},
	}

	cmd.Flags().IntVar(&prNumber, "pr", 0, "PR number (auto-detected from current branch if not set)")
	cmd.Flags().StringVar(&intervalStr, "interval", "30s", "Poll interval (e.g. 30s, 1m)")

	return cmd
}

// watchLoop polls for new comments and reviews, printing new items to stdout.
// It stops on SIGINT (via ctx cancellation), applies exponential backoff after repeated errors,
// and caps knownIDs to avoid unbounded memory growth.
func (a *app) watchLoop(ctx context.Context, prNumber int, interval time.Duration) error {
	// get initial state
	known, err := a.commentService.List(prNumber)
	if err != nil {
		return fmt.Errorf("commentService.List [pr=%d]: %w", prNumber, err)
	}

	// build set of known comment/review IDs (capped ring-buffer style)
	knownIDs := make(map[int64]bool)
	for _, c := range known.IssueComments {
		knownIDs[c.DatabaseID] = true
	}
	for _, r := range known.Reviews {
		knownIDs[r.DatabaseID] = true
	}

	fmt.Printf("watching PR #%d (interval %s, %d existing items)\n", prNumber, interval, len(knownIDs))

	var consecErrors int
	backoff := interval

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(backoff):
		}

		current, err := a.commentService.List(prNumber)
		if err != nil {
			consecErrors++
			fmt.Printf("error polling (attempt %d/%d): %v\n", consecErrors, watchMaxConsecErrors, err)
			if consecErrors >= watchMaxConsecErrors {
				return fmt.Errorf("commentService.List [pr=%d]: %d consecutive errors, last: %w",
					prNumber, consecErrors, err)
			}
			backoff = watchErrorBackoffBase * time.Duration(consecErrors)
			continue
		}

		// reset on success
		consecErrors = 0
		backoff = interval

		// detect new issue comments
		for _, c := range current.IssueComments {
			if !knownIDs[c.DatabaseID] {
				addKnownID(knownIDs, c.DatabaseID)
				fmt.Printf("\n--- new comment #%d by @%s at %s ---\n%s\n",
					c.DatabaseID, c.Author, c.CreatedAt, c.Body)
			}
		}

		// detect new reviews
		for _, r := range current.Reviews {
			if !knownIDs[r.DatabaseID] {
				addKnownID(knownIDs, r.DatabaseID)
				fmt.Printf("\n--- new review #%d (%s) by @%s at %s ---\n",
					r.DatabaseID, r.State, r.Author, r.CreatedAt)
				if r.Body != "" {
					fmt.Println(r.Body)
				}
				if r.CommentCount > 0 {
					fmt.Printf("(%d inline comments — run: gh-contribute review %d --pr %d)\n",
						r.CommentCount, r.DatabaseID, prNumber)
				}
			}
		}
	}
}

// addKnownID adds id to the set, evicting the oldest entry when the cap is reached.
// This prevents knownIDs from growing unboundedly over long-running watch sessions.
func addKnownID(ids map[int64]bool, id int64) {
	if len(ids) >= watchMaxKnownIDs {
		// evict an arbitrary entry (map iteration order is random in Go)
		for old := range ids {
			delete(ids, old)
			break
		}
	}
	ids[id] = true
}
