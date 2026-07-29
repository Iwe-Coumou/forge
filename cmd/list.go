package cmd

import (
	"fmt"

	"github.com/Iwe-Coumou/forge/v2/internal/forger"
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
		id := t.ID()
		if t.NotImplemented != "" {
			id += " (wip)"
		}

		if verbose {
			color.Blue("%s\n %s\n", id, t.Long)
			if t.NotImplemented != "" {
				color.Yellow(" not implemented yet: %s\n", t.NotImplemented)
			}
			fmt.Println()
		} else {
			color.Blue("%-26s %s\n", id, t.Short)
		}
	}

	return nil
}
