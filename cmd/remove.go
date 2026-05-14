package cmd

import (
	"github.com/matdexir/c3vm/internal/c3vm"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <version>",
	Short: "Remove an installed version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vm, err := c3vm.New()
		if err != nil {
			return err
		}
		return vm.Remove(args[0])
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
