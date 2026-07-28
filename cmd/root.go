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
	Use:   "forge",
	Short: "Forge is a scaffolding CLI to create Go project.",
	Long:  "Long description",
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
