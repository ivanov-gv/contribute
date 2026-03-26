package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ivanov-gv/gh-contribute/internal/service/issue"
)

func (a *app) newIssueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue <number>",
		Short: "Show issue details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			number, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid issue number '%s': %w", args[0], err)
			}

			info, err := a.issueService.Get(number)
			if err != nil {
				return fmt.Errorf("issueService.Get [number=%d]: %w", number, err)
			}

			fmt.Print(info.Format())
			return nil
		},
	}

	return cmd
}

const defaultIssueListLimit = 20

func (a *app) newIssuesCmd() *cobra.Command {
	var labelFlag string
	var limit int

	cmd := &cobra.Command{
		Use:   "issues",
		Short: "List open issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			var labels []string
			if labelFlag != "" {
				labels = strings.Split(labelFlag, ",")
			}

			items, err := a.issueService.List(limit, labels)
			if err != nil {
				return fmt.Errorf("issueService.List: %w", err)
			}

			fmt.Print(issue.FormatList(items))
			return nil
		},
	}

	cmd.Flags().StringVar(&labelFlag, "label", "", "Filter by label (comma-separated for multiple)")
	cmd.Flags().IntVar(&limit, "limit", defaultIssueListLimit, "Maximum number of issues to return")
	return cmd
}
