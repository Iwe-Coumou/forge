package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Iwe-Coumou/forge/internal/config"
	"github.com/Iwe-Coumou/forge/internal/forger"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var projectPath string

var newCmd = &cobra.Command{
	Use:   "new [template] [project-name]",
	Short: "Scaffold a new project from a template",
	Args:  cobra.ExactArgs(2),
	RunE:  runNew,
}

func init() {
	newCmd.Flags().StringVarP(&projectPath, "path", "p", "", "Go module path (defaults to current directory name)")
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	template := args[0]
	projectName := args[1]

	outputDir, err := resolveOutputDir(projectName)
	if err != nil {
		return fmt.Errorf("resolving output directory: %w", err)
	}

	modulePath, err := resolveModulePath(projectName)
	if err != nil {
		return fmt.Errorf("resolving module path: %w", err)
	}

	projectParams := &forger.Project{
		Name:       projectName,
		ModulePath: modulePath,
		OutputDir:  outputDir,
		Template:   template,
	}

	if err := forger.Forge(projectParams, verbose); err != nil {
		return fmt.Errorf("scaffolding: %w", err)
	}

	color.Green("Forged succesfully")
	return nil
}

func resolveOutputDir(projectName string) (string, error) {
	baseDir := projectPath
	if baseDir == "" {
		var err error
		baseDir, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	return filepath.Join(baseDir, projectName), nil
}

func resolveModulePath(projectName string) (string, error) {
	cfg, err := config.Load()
	if err != nil {
		return "", err
	}

	if cfg.BaseModule == "" {
		return projectName, nil
	}
	return cfg.BaseModule + "/" + projectName, nil
}
