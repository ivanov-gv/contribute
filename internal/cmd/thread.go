package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func (a *app) newThreadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "thread <thread-id>",
		Short: "Show all comments in a thread across all reviews",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			threadID, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid thread ID '%s': %w", args[0], err)
			}

			prNumber, _ := cmd.Flags().GetInt("pr")
			number, err := a.resolvePR(prNumber)
			if err != nil {
				return fmt.Errorf("resolvePR [pr=%d]: %w", prNumber, err)
			}

			t, err := a.threadService.Get(number, threadID)
			if err != nil {
				return fmt.Errorf("threadService.Get [pr=%d, thread=%d]: %w", number, threadID, err)
			}

			showHidden, _ := cmd.Flags().GetBool("show-hidden")
			fmt.Print(t.Format(showHidden))
			return nil
		},
	}

	cmd.Flags().Int("pr", 0, "PR number (auto-detected from current branch if not set)")
	cmd.Flags().Bool("show-hidden", false, "Show content of hidden/minimized comments")
	return cmd
}
