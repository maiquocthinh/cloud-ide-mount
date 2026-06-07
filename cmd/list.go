package cmd

import (
	"fmt"

	"cloud-ide-mount/internal/codespace"
	"cloud-ide-mount/internal/ui"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show all codespaces and state",
	RunE: func(_ *cobra.Command, _ []string) error {
		all, err := codespace.List()
		if err != nil {
			return err
		}
		if len(all) == 0 {
			fmt.Println("No codespaces found.")
			return nil
		}
		ui.ShowCsList(all)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
