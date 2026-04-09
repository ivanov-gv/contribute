package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// outputFormat returns the --format flag value from the command
func outputFormat(cmd *cobra.Command) string {
	f, _ := cmd.Flags().GetString("format")
	return f
}

// printOutput prints data as markdown (default) or JSON based on --format flag
func printOutput(cmd *cobra.Command, data interface{ Format() string }, jsonData interface{}) error {
	if outputFormat(cmd) == "json" {
		return printJSON(jsonData)
	}
	fmt.Print(data.Format())
	return nil
}

// printJSON marshals data to JSON and prints it
func printJSON(data interface{}) error {
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("json.Marshal: %w", err)
	}
	fmt.Println(string(out))
	return nil
}
