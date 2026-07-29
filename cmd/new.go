package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Iwe-Coumou/forge/v2/internal/config"
	"github.com/Iwe-Coumou/forge/v2/internal/forger"
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
	newCmd.Flags().StringVarP(&projectPath, "path", "p", "", "Directory to scaffold into (defaults to the current directory)")
	newCmd.Flags().BoolVarP(&gitInit, "git", "g", false, "Initialize git repository in the new project")
	newCmd.Flags().StringToStringVar(&overrides, "set", nil, "Override a language-specific value, e.g. --set module_path=github.com/me/x")
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	lang, tmplName, err := forger.ParseTemplateID(cfg, args[0])
	if err != nil {
		return err
	}
	projectName := args[1]

	outputDir, err := resolveOutputDir(projectName)
	if err != nil {
		return fmt.Errorf("resolving output directory: %w", err)
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

	// The -g flag defaults to false, so "not passed" and "passed as false"
	// look identical — Changed() is what lets config supply the default.
	useGit := gitInit
	if !cmd.Flags().Changed("git") {
		useGit = cfg.GitInit
	}

	postProcessOptions := &forger.PostProcessOptions{GitInit: useGit}
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
