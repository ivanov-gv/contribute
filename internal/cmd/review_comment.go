package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (a *app) newReviewCommentCmd() *cobra.Command {
	var prNumber int
	var filePath string
	var line int
	var side string

	cmd := &cobra.Command{
		Use:   "review-comment <body>",
		Short: "Post an inline review comment on a specific file and line",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			body := args[0]

			// resolve PR number
			number, err := a.resolvePR(prNumber)
			if err != nil {
				return fmt.Errorf("resolvePR [pr=%d]: %w", prNumber, err)
			}

			// get head commit SHA from PR
			info, err := a.prService.Get(number)
			if err != nil {
				return fmt.Errorf("prService.Get [number=%d]: %w", number, err)
			}

			created, err := a.commentService.PostInlineComment(number, info.HeadCommitSHA, filePath, body, line, side)
			if err != nil {
				return fmt.Errorf("commentService.PostInlineComment [pr=%d, path='%s', line=%d]: %w", number, filePath, line, err)
			}

			fmt.Printf("posted inline comment #%d on %s:%d in PR #%d\n", created.DatabaseID, filePath, line, number)
			return nil
		},
	}

	cmd.Flags().IntVar(&prNumber, "pr", 0, "PR number (auto-detected from current branch if not set)")
	cmd.Flags().StringVar(&filePath, "file", "", "File path relative to repo root")
	cmd.Flags().IntVar(&line, "line", 0, "Line number in the diff")
	cmd.Flags().StringVar(&side, "side", "RIGHT", "Diff side: RIGHT (default) or LEFT")

	_ = cmd.MarkFlagRequired("file")
	_ = cmd.MarkFlagRequired("line")

	return cmd
}
