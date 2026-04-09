package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// validReviewEvents lists the accepted --event values
var validReviewEvents = []string{"APPROVE", "REQUEST_CHANGES", "COMMENT"}

func (a *app) newSubmitReviewCmd() *cobra.Command {
	var prNumber int
	var event string
	var body string

	cmd := &cobra.Command{
		Use:   "submit-review",
		Short: "Submit a review on a PR",
		Long:  fmt.Sprintf("Submit a review with an event type.\nValid events: %s", strings.Join(validReviewEvents, ", ")),
		RunE: func(cmd *cobra.Command, args []string) error {
			// validate event
			event = strings.ToUpper(event)
			if !isValidEvent(event) {
				return fmt.Errorf("invalid event '%s', valid: %v", event, validReviewEvents)
			}

			// resolve PR number
			number, err := a.resolvePR(prNumber)
			if err != nil {
				return fmt.Errorf("resolvePR [pr=%d]: %w", prNumber, err)
			}

			reviewID, err := a.commentService.SubmitReview(number, event, body)
			if err != nil {
				return fmt.Errorf("commentService.SubmitReview [pr=%d, event='%s']: %w", number, event, err)
			}

			fmt.Printf("submitted %s review #%d on PR #%d\n", event, reviewID, number)
			return nil
		},
	}

	cmd.Flags().IntVar(&prNumber, "pr", 0, "PR number (auto-detected from current branch if not set)")
	cmd.Flags().StringVar(&event, "event", "", "Review event: APPROVE, REQUEST_CHANGES, or COMMENT")
	cmd.Flags().StringVar(&body, "body", "", "Optional review body text")

	_ = cmd.MarkFlagRequired("event")

	return cmd
}

func isValidEvent(event string) bool {
	for _, e := range validReviewEvents {
		if e == event {
			return true
		}
	}
	return false
}
