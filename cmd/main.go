package main

import (
	"os"

	"github.com/mdb/gh-dispatch/internal/dispatch"
)

// version's value is passed in at build time.
var version string

func main() {
	rootCmd := dispatch.NewCmdRoot(version)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
