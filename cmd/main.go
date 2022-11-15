package main

import (
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/mdb/gh-dispatch/internal/dispatch"
	"github.com/spf13/cobra"
)

// version's value is passed in at build time.
var version string

func main() {
	rootCmd := &cobra.Command{
		Use:   "gh dispatch",
		Short: "Send a GitHub dispatch event and watch the resulting GitHub Actions run",
		Long: heredoc.Doc(`
		Send a workflow_dispatch or repository_dispatch event and watch the resulting
		GitHub Actions run.
	`),
		SilenceUsage: true,
		Version:      version,
	}

	// TODO: how to make this required?
	rootCmd.PersistentFlags().StringP("repo", "R", "", "The targeted repository's full name (in 'owner/repo' format)")

	repositoryCmd := dispatch.NewCmdRepository()
	rootCmd.AddCommand(repositoryCmd)

	workflowCmd := dispatch.NewCmdWorkflow()
	rootCmd.AddCommand(workflowCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
