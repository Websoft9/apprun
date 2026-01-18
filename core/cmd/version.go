package cmd

import (
	"fmt"

	"apprun/pkg/version"

	"github.com/spf13/cobra"
)

var (
	shortVersion bool
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:     "version",
	GroupID: "server",
	Short:   "Display version information",
	Long: `Display version information including build details.

Examples:
  apprun version           # Full version info
  apprun version --short   # Just version number
`,
	Run: runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVar(&shortVersion, "short", false, "display only the version number")
}

func runVersion(cmd *cobra.Command, args []string) {
	if shortVersion {
		fmt.Println(version.Short())
	} else {
		fmt.Println(version.Info())
	}
}
