package cmd

import (
	"fmt"
	"log/slog"

	"github.com/matdexir/c3vm/internal/c3vm"
	"github.com/spf13/cobra"
)

var defaultCmd = &cobra.Command{
	Use:   "default [version]",
	Short: "Get or set the default version",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vm, err := c3vm.New()
		if err != nil {
			return err
		}

		if len(args) == 0 {
			def, err := vm.Default()
			if err != nil {
				return err
			}
			if def == "" {
				slog.Warn("no default version set")
				return nil
			}
			fmt.Println(def)
			return nil
		}

		return vm.SetDefault(args[0])
	},
}

func init() {
	rootCmd.AddCommand(defaultCmd)
}
