package cmd

import (
	"github.com/mdb/gh-dispatch/dispatch"
)

func init() {
	workflowCmd := dispatch.NewCmdWorkflow()
	rootCmd.AddCommand(workflowCmd)
}
