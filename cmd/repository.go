package cmd

import (
	"github.com/spf13/cobra"
)

// repositoryCmd represents the repository subcommand
var repositoryCmd = &cobra.Command{
	Use:     "repository",
	Short:   `The 'repository' subcommand triggers repository dispatch events`,
	Long:    `The 'repository' subcommand triggers repository dispatch events`,
	Example: `TODO`,
}

func init() {
	rootCmd.AddCommand(repositoryCmd)
}
