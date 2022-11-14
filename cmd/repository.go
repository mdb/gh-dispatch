package cmd

import "github.com/mdb/gh-dispatch/dispatch"

var (
	repositoryEventType     string
	repositoryClientPayload string
	repositoryWorkflow      string
)

// repositoryCmd represents the repository subcommand
var repositoryCmd = dispatch.NewCmdRepository()

func init() {
	repositoryCmd.Flags().StringVarP(&repositoryEventType, "event-type", "e", "", "The repository dispatch event type.")
	repositoryCmd.MarkFlagRequired("event-type")
	repositoryCmd.Flags().StringVarP(&repositoryClientPayload, "client-payload", "p", "", "The repository dispatch event client payload JSON string.")
	repositoryCmd.MarkFlagRequired("client-payload")
	repositoryCmd.Flags().StringVarP(&repositoryWorkflow, "workflow", "w", "", "The resulting GitHub Actions workflow name.")
	repositoryCmd.MarkFlagRequired("workflow")

	rootCmd.AddCommand(repositoryCmd)
}
