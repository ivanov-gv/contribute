package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func (a *app) newCommentsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comments [comment-id]",
		Short: "List comments on a PR, or show a single comment by ID",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// resolve PR number
			prNumber, _ := cmd.Flags().GetInt("pr")
			number, err := a.resolvePR(prNumber)
			if err != nil {
				return fmt.Errorf("resolvePR [pr=%d]: %w", prNumber, err)
			}

			showHidden, _ := cmd.Flags().GetBool("show-hidden")

			result, err := a.commentService.List(number)
			if err != nil {
				return fmt.Errorf("commentService.List [pr=%d]: %w", number, err)
			}

			// filter by comment ID if provided — single-item lookup always shows full content
			if len(args) > 0 {
				commentID, err := strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid comment ID '%s': %w", args[0], err)
				}
				filtered := result.FilterByID(commentID)
				if filtered == nil {
					return fmt.Errorf("comment #%d not found", commentID)
				}
				if outputFormat(cmd) == "json" {
					return printJSON(filtered)
				}
				// single-item: always show hidden, normalize to single trailing newline
				output := filtered.Format(true)
				fmt.Print(strings.TrimRight(output, "\n") + "\n")
				return nil
			}

			if outputFormat(cmd) == "json" {
				return printJSON(result)
			}
			fmt.Print(result.Format(showHidden))
			return nil
		},
	}

	cmd.Flags().Int("pr", 0, "PR number (auto-detected from current branch if not set)")
	cmd.Flags().Bool("show-hidden", false, "Show content of hidden/minimized comments")
	return cmd
}
