package cmd

import (
	"fmt"

	"github.com/matdexir/c3vm/internal/c3vm"
	"github.com/spf13/cobra"
)

var whichCmd = &cobra.Command{
	Use:   "which [version]",
	Short: "Show the path to the c3c binary",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vm, err := c3vm.New()
		if err != nil {
			return err
		}

		version := ""
		if len(args) > 0 {
			version = args[0]
		}

		path, err := vm.Which(version)
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(whichCmd)
}
