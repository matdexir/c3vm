package cmd

import (
	"github.com/matdexir/c3vm/internal/c3vm"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:     "install <version|latest>",
	Aliases: []string{"i"},
	Short:   "Install a specific version or the latest release",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vm, err := c3vm.New()
		if err != nil {
			return err
		}
		return vm.Install(args[0])
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
