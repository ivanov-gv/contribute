package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

const issueCommentArgCount = 2 // <issue-number> <body>

func (a *app) newIssueCommentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue-comment <issue-number> <body>",
		Short: "Post a comment on an issue",
		Args:  cobra.ExactArgs(issueCommentArgCount),
		RunE: func(cmd *cobra.Command, args []string) error {
			issueNumber, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid issue number '%s': %w", args[0], err)
			}
			body := args[1]

			// reuse comment service — GitHub's issue comment API works for both issues and PRs
			created, err := a.commentService.Post(issueNumber, body)
			if err != nil {
				return fmt.Errorf("commentService.Post [issue=%d]: %w", issueNumber, err)
			}

			fmt.Printf("posted comment #%d on issue #%d\n", created.DatabaseID, issueNumber)
			return nil
		},
	}

	return cmd
}
