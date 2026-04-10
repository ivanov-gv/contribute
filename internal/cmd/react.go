package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/ivanov-gv/contribute/internal/service/reaction"
)

const reactArgCount = 2 // <id> <reaction>

func (a *app) newReactCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "react",
		Short: "Add a reaction to a comment or review",
		Long: fmt.Sprintf(
			"Add a reaction to a PR comment or review body.\nValid reactions: %v\n\nSubcommands:\n  comment <id> <reaction>  — react to a review comment (inline)\n  review  <id> <reaction>  — react to a review body\n  issue-comment <id> <reaction> — react to an issue/PR top-level comment",
			reaction.ValidReactions,
		),
	}

	cmd.AddCommand(
		a.newReactCommentCmd(),
		a.newReactReviewCmd(),
		a.newReactIssueCommentCmd(),
	)

	return cmd
}

// newReactCommentCmd reacts to an inline review comment (pull request review comment)
func (a *app) newReactCommentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "comment <comment-id> <reaction>",
		Short: "Add a reaction to an inline review comment",
		Args:  cobra.ExactArgs(reactArgCount),
		RunE: func(cmd *cobra.Command, args []string) error {
			commentID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid comment ID '%s': %w", args[0], err)
			}
			reactionContent := args[1]

			if err := a.reactionService.AddToReviewComment(commentID, reactionContent); err != nil {
				return fmt.Errorf("reactionService.AddToReviewComment [comment=%d, reaction='%s']: %w", commentID, reactionContent, err)
			}

			fmt.Printf("added '%s' reaction to comment %d\n", reactionContent, commentID)
			return nil
		},
	}
}

// newReactReviewCmd reacts to a PR review body using GraphQL (no REST endpoint exists)
func (a *app) newReactReviewCmd() *cobra.Command {
	var prNumber int

	cmd := &cobra.Command{
		Use:   "review <review-id> <reaction>",
		Short: "Add a reaction to a PR review body",
		Args:  cobra.ExactArgs(reactArgCount),
		RunE: func(cmd *cobra.Command, args []string) error {
			reviewID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid review ID '%s': %w", args[0], err)
			}
			reactionContent := args[1]

			// resolve PR number — required to locate the review via GraphQL
			number, err := a.resolvePR(prNumber)
			if err != nil {
				return fmt.Errorf("resolvePR [pr=%d]: %w", prNumber, err)
			}

			if err := a.reactionService.AddToReview(number, reviewID, reactionContent); err != nil {
				return fmt.Errorf("reactionService.AddToReview [pr=%d, review=%d, reaction='%s']: %w", number, reviewID, reactionContent, err)
			}

			fmt.Printf("added '%s' reaction to review %d\n", reactionContent, reviewID)
			return nil
		},
	}

	cmd.Flags().IntVar(&prNumber, "pr", 0, "PR number (auto-detected from current branch if not set)")
	return cmd
}

// newReactIssueCommentCmd reacts to a top-level issue or PR comment
func (a *app) newReactIssueCommentCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "issue-comment <comment-id> <reaction>",
		Short: "Add a reaction to an issue or PR top-level comment",
		Args:  cobra.ExactArgs(reactArgCount),
		RunE: func(cmd *cobra.Command, args []string) error {
			commentID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid comment ID '%s': %w", args[0], err)
			}
			reactionContent := args[1]

			if err := a.reactionService.AddToIssueComment(commentID, reactionContent); err != nil {
				return fmt.Errorf("reactionService.AddToIssueComment [comment=%d, reaction='%s']: %w", commentID, reactionContent, err)
			}

			fmt.Printf("added '%s' reaction to issue comment %d\n", reactionContent, commentID)
			return nil
		},
	}
}
