package cmd

import "github.com/mdb/gh-dispatch/dispatch"

func init() {
	repositoryCmd := dispatch.NewCmdRepository()
	rootCmd.AddCommand(repositoryCmd)
}
