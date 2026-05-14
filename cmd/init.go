package cmd

import (
	"fmt"

	"github.com/matdexir/c3vm/internal/c3vm"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Print shell setup instructions for PATH",
	RunE: func(cmd *cobra.Command, args []string) error {
		vm, err := c3vm.New()
		if err != nil {
			return err
		}

		if vm.BinInPath() {
			fmt.Println("~/.c3vm/bin is already in your PATH.")
			return nil
		}

		fmt.Println("Add ~/.c3vm/bin to your PATH by running:")
		fmt.Println()
		fmt.Printf("  %s\n", vm.InitScript())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
