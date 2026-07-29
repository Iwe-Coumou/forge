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
var overrides map[string]string

var newCmd = &cobra.Command{
	Use:   "new [language/template] [project-name]",
	Short: "Forge a new project from a template",
	Args:  cobra.ExactArgs(2),
	RunE:  runNew,
}

func init() {
	newCmd.Flags().StringVarP(&projectPath, "path", "p", "", "Output directory for the project (defaults to current directory name)")
	newCmd.Flags().BoolVarP(&gitInit, "git", "g", false, "Initialize git repository in the new project")
	newCmd.Flags().StringToStringVar(&overrides, "set", nil, "Override a language-specific value, e.g. --set module_path=github.com/me/x")
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	lang, tmplName, err := forger.ParseTemplateID(args[0])
	if err != nil {
		return err
	}
	projectName := args[1]

	outputDir, err := resolveOutputDir(projectName)
	if err != nil {
		return fmt.Errorf("resolving output directory: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	projectParams := &forger.Project{
		Name:      projectName,
		OutputDir: outputDir,
		Language:  lang,
		Template:  tmplName,
		Overrides: overrides,
	}

	if err := forger.Forge(projectParams, cfg, verbose); err != nil {
		return fmt.Errorf("scaffolding: %w", err)
	}
	color.Green("Forged successfully\n")

	postProcessOptions := &forger.PostProcessOptions{GitInit: gitInit}
	if err := forger.PostProcess(projectParams, postProcessOptions, verbose); err != nil {
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
