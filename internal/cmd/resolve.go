package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func (a *app) newResolveCmd() *cobra.Command {
	var prNumber int

	cmd := &cobra.Command{
		Use:   "resolve <thread-id>",
		Short: "Resolve a review thread",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			threadID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid thread ID '%s': %w", args[0], err)
			}

			// resolve PR number
			number, err := a.resolvePR(prNumber)
			if err != nil {
				return fmt.Errorf("resolvePR [pr=%d]: %w", prNumber, err)
			}

			if err := a.threadService.Resolve(number, threadID); err != nil {
				return fmt.Errorf("threadService.Resolve [pr=%d, thread=%d]: %w", number, threadID, err)
			}

			fmt.Printf("resolved thread #%d on PR #%d\n", threadID, number)
			return nil
		},
	}

	cmd.Flags().IntVar(&prNumber, "pr", 0, "PR number (auto-detected from current branch if not set)")
	return cmd
}

func (a *app) newUnresolveCmd() *cobra.Command {
	var prNumber int

	cmd := &cobra.Command{
		Use:   "unresolve <thread-id>",
		Short: "Unresolve a review thread",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			threadID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid thread ID '%s': %w", args[0], err)
			}

			number, err := a.resolvePR(prNumber)
			if err != nil {
				return fmt.Errorf("resolvePR [pr=%d]: %w", prNumber, err)
			}

			if err := a.threadService.Unresolve(number, threadID); err != nil {
				return fmt.Errorf("threadService.Unresolve [pr=%d, thread=%d]: %w", number, threadID, err)
			}

			fmt.Printf("unresolved thread #%d on PR #%d\n", threadID, number)
			return nil
		},
	}

	cmd.Flags().IntVar(&prNumber, "pr", 0, "PR number (auto-detected from current branch if not set)")
	return cmd
}
