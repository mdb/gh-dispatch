package cmd

import (
	"github.com/mdb/gh-dispatch/dispatch"
)

var (
	workflowInputs string
	workflowName   string
	workflowRef    string
)

// workflowCmd represents the workflow subcommand
var workflowCmd = dispatch.NewCmdWorkflow()

func init() {
	// TODO: how does the 'gh run' command represent inputs?
	// Is it worth better emulating its interface?
	workflowCmd.Flags().StringVarP(&workflowInputs, "inputs", "i", "", "The workflow dispatch inputs JSON string.")
	workflowCmd.MarkFlagRequired("inputs")
	// TODO: how does the 'gh run' command represent workflow?
	// Is it worth better emulating its interface?
	workflowCmd.Flags().StringVarP(&workflowName, "workflow", "w", "", "The resulting GitHub Actions workflow name.")
	workflowCmd.MarkFlagRequired("workflow")
	// TODO: how does the 'gh run' command represent ref?
	// Is it worth better emulating its interface?
	workflowCmd.Flags().StringVarP(&workflowRef, "ref", "f", "main", "The git reference for the workflow. Can be a branch or tag name.")

	rootCmd.AddCommand(workflowCmd)
}
