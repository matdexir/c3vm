package cmd

import (
	"fmt"
	"log/slog"

	"github.com/matdexir/c3vm/internal/c3vm"
	"github.com/spf13/cobra"
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the currently active version",
	RunE: func(cmd *cobra.Command, args []string) error {
		vm, err := c3vm.New()
		if err != nil {
			return err
		}
		current, err := vm.Current()
		if err != nil {
			return err
		}
		if current == "" {
			slog.Warn("no active version set")
			return nil
		}
		fmt.Println(current)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(currentCmd)
}
