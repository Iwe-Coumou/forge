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

// version is set at build time via:
//
//	go build -ldflags "-X github.com/Iwe-Coumou/forge/cmd.version=v1.0.0"
var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "forge",
	Short: "Forge is a scaffolding CLI to create coding project.",
	Long: `Forge is a scaffolding CLI for coding projects.

Run "forge init" once to set your default module base path, "forge list"
to see the available templates, and "forge new <language/template> <name>" to
generate a project. Forge renders the template, tidies its dependencies,
formats the result, and can optionally initialize a git repository.`,
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(banner)
		color.HiBlack("  %s\n\n", version)
		cmd.Help()
	},
	SilenceErrors: true,
	SilenceUsage:  true,
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
