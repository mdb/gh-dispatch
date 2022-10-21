package cmd

import (
	"github.com/spf13/cobra"
)

// workflowCmd represents the workflow subcommand
var workflowCmd = &cobra.Command{
	Use:     "workflow",
	Short:   `The 'workflow' subcommand triggers workflow dispatch events`,
	Long:    `The 'workflow' subcommand triggers workflow dispatch events`,
	Example: `TODO`,
}

func init() {
	rootCmd.AddCommand(workflowCmd)
}
