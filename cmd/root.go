package cmd

import (
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/spf13/cobra"
)

// rootCmd is the root command.
var rootCmd = &cobra.Command{
	Use:   "gh dispatch",
	Short: "Send a GitHub dispatch event and watch the resulting GitHub Actions run",
	Long: heredoc.Doc(`
		Send a workflow_dispatch or repository_dispatch event and watch the resulting
		GitHub Actions run.
	`),
	SilenceUsage: true,
}

func init() {
	// TODO: how to make this required?
	rootCmd.PersistentFlags().StringP("repo", "R", "", "The targeted repository's full name (in 'owner/repo' format)")
}

// Execute executes the root command.
func Execute(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
