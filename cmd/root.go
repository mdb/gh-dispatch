package cmd

import (
	"github.com/spf13/cobra"
)

// rootCmd is the root command.
var rootCmd = &cobra.Command{
	Use:     "gh dispatch",
	Short:   "gh dispatch: Trigger a GitHub dispatch event and watch the resulting GitHub Actions run",
	Long:    "gh dispatch: Trigger a GitHub dispatch event and watch the resulting GitHub Actions run",
	Example: "TODO",
}

func init() {
	rootCmd.PersistentFlags().BoolP("watch", "w", true, "Watch the resulting GitHub Actions run")
}

// Execute executes the root command.
func Execute(version string) {
	rootCmd.Version = version
	cobra.CheckErr(rootCmd.Execute())
}
