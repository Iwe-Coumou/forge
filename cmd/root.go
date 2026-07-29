package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const banner = `
░██████████                                          
░██                                                  
░██         ░███████  ░██░████  ░████████  ░███████  
░█████████ ░██    ░██ ░███     ░██    ░██ ░██    ░██ 
░██        ░██    ░██ ░██      ░██    ░██ ░█████████ 
░██        ░██    ░██ ░██      ░██   ░███ ░██        
░██         ░███████  ░██       ░█████░██  ░███████  
                                      ░██            
                                ░███████            
`

var verbose bool

var rootCmd = &cobra.Command{
	Use:     "forge",
	Short: "Forge is a scaffolding CLI to create Go project.",
	Long: `Forge is a scaffolding CLI for Go projects.

Run "forge init" once to set your default module base path, "forge list"
to see the available templates, and "forge new <template> <name>" to
generate a project. Forge renders the template, tidies its dependencies,
formats the result, and can optionally initialize a git repository.`,
	Version: "1.0.0",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(banner)
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		color.Red(err.Error())
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
}
