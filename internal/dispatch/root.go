package dispatch

import (
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/go-gh/v2/pkg/repository"
	"github.com/spf13/cobra"
)

func NewCmdRoot(version string) *cobra.Command {
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

	defaultRepo := ""
	currentRepo, err := repository.Current()
	if err == nil {
		defaultRepo = fmt.Sprintf("%s/%s/%s", currentRepo.Host, currentRepo.Owner, currentRepo.Name)
	}

	var repo string
	rootCmd.PersistentFlags().StringVarP(&repo, "repo", "R", defaultRepo, "The targeted repository's full name")

	repositoryCmd := NewCmdRepository()
	rootCmd.AddCommand(repositoryCmd)

	workflowCmd := NewCmdWorkflow()
	rootCmd.AddCommand(workflowCmd)

	return rootCmd
}
