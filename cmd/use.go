package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/matdexir/c3vm/internal/c3vm"
	"github.com/spf13/cobra"
)

var useYes bool

var useCmd = &cobra.Command{
	Use:   "use <version>",
	Short: "Switch to a specific version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vm, err := c3vm.New()
		if err != nil {
			return err
		}

		version := args[0]

		tag, err := vm.ResolveVersion(version)
		if err != nil {
			return err
		}

		// Check if the version is installed
		dir := vm.VersionDir(tag)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if !useYes {
				if !promptYesNo(fmt.Sprintf("Version %s is not installed. Download it now?", tag)) {
					return fmt.Errorf("version %s is not installed", tag)
				}
			}

			if err := vm.Install(version); err != nil {
				return err
			}
		}

		return vm.Use(tag)
	},
}

func init() {
	useCmd.Flags().BoolVarP(&useYes, "yes", "y", false, "Automatically confirm download")
	rootCmd.AddCommand(useCmd)
}

func promptYesNo(question string) bool {
	fmt.Printf("%s [y/N] ", question)

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return false
	}

	answer := strings.TrimSpace(scanner.Text())
	return answer == "y" || answer == "Y" || answer == "yes" || answer == "Yes" || answer == "YES"
}
