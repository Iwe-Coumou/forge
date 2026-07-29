package cmd

import (
	"fmt"

	"github.com/Iwe-Coumou/forge/internal/forger"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates",
	Args:  cobra.NoArgs,
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	templates, err := forger.ListTemplates()
	if err != nil {
		return fmt.Errorf("listing templates: %w", err)
	}

	for _, t := range templates {
		if verbose {
			color.Blue("%s\n %s\n\n", t.Name, t.Long)
		} else {
			color.Blue("%-15s %s\n", t.Name, t.Short)
		}
	}

	return nil
}
