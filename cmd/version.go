package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the c3vm version, set at build time via
// -ldflags "-X github.com/matdexir/c3vm/cmd.Version=<version>".
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the c3vm version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
