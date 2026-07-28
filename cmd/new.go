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
var gitInit bool

var newCmd = &cobra.Command{
	Use:   "new [template] [project-name]",
	Short: "Forge a new project from a template",
	Args:  cobra.ExactArgs(2),
	RunE:  runNew,
}

func init() {
	newCmd.Flags().StringVarP(&projectPath, "path", "p", "", "Go module path (defaults to current directory name)")
	newCmd.Flags().BoolVarP(&gitInit, "git", "g", false, "Initialize git repository in the new project")
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
	color.Green("Forged successfully\n")

	postProcessOptions := &forger.PostProcessOptions{GitInit: gitInit}
	if err := forger.PostProcess(projectParams.OutputDir, postProcessOptions, verbose); err != nil {
		return fmt.Errorf("post-processing: %w", err)
	}
	color.Green("Post-processing successful")

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
