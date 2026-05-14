package cmd

import (
	"fmt"

	"github.com/matdexir/c3vm/internal/c3vm"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed versions",
	RunE: func(cmd *cobra.Command, args []string) error {
		vm, err := c3vm.New()
		if err != nil {
			return err
		}
		versions, err := vm.List()
		if err != nil {
			return err
		}

		current, _ := vm.Current()

		if len(versions) == 0 {
			fmt.Println("No versions installed")
			return nil
		}

		for _, v := range versions {
			mark := " "
			if v == current {
				mark = "*"
			}
			fmt.Printf(" %s %s\n", mark, v)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
