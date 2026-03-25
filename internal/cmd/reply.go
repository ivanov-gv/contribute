package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func (a *app) newReplyCmd() *cobra.Command {
	var prNumber int

	cmd := &cobra.Command{
		Use:   "reply <comment-id> <body>",
		Short: "Reply to a review comment in-thread",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			commentID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid comment ID '%s': %w", args[0], err)
			}
			body := args[1]

			// resolve PR number
			number, err := a.resolvePR(prNumber)
			if err != nil {
				return err
			}

			created, err := a.commentService.ReplyToReviewComment(number, commentID, body)
			if err != nil {
				return fmt.Errorf("commentService.ReplyToReviewComment [pr=%d, comment=%d]: %w", number, commentID, err)
			}

			fmt.Printf("posted reply #%d to comment #%d on PR #%d\n", created.DatabaseID, commentID, number)
			return nil
		},
	}

	cmd.Flags().IntVar(&prNumber, "pr", 0, "PR number (auto-detected from current branch if not set)")
	return cmd
}
