package cmd

import (
	"fmt"
	"log/slog"

	"github.com/matdexir/c3vm/internal/c3vm"
	"github.com/spf13/cobra"
)

var listRemoteCmd = &cobra.Command{
	Use:     "list-remote",
	Aliases: []string{"lr"},
	Short:   "List available versions from GitHub",
	RunE: func(cmd *cobra.Command, args []string) error {
		vm, err := c3vm.New()
		if err != nil {
			return err
		}
		tags, err := vm.ListRemote()
		if err != nil {
			return err
		}
		if len(tags) == 0 {
			slog.Warn("no releases found")
			return nil
		}
		for _, tag := range tags {
			fmt.Println(tag)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listRemoteCmd)
}
