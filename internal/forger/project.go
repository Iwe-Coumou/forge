package forger

import (
	"fmt"
	"os"
)

type Project struct {
	Name       string
	ModulePath string
	OutputDir  string
	Template   string
}

func (p *Project) validate() error {
	if p.Name == "" {
		return fmt.Errorf("project name is required")
	}

	if p.ModulePath == "" {
		return fmt.Errorf("project module path is required")
	}

	if p.OutputDir == "" {
		return fmt.Errorf("project output directory is required")
	}

	if p.Template == "" {
		return fmt.Errorf("project template is required")
	}

	if !templateExists(p.Template) {
		return fmt.Errorf("unknown template %q", p.Template)
	}

	if info, err := os.Stat(p.OutputDir); err == nil && info.IsDir() {
		entries, err := os.ReadDir(p.OutputDir)
		if err != nil {
			return fmt.Errorf("checking output directory: %w", err)
		}
		if len(entries) > 0 {
			return fmt.Errorf("output directory %s already exists and is not empty", p.OutputDir)
		}
	}

	return nil
}

func templateExists(templateName string) bool {
	entries, err := templateFS.ReadDir("templates")
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() && e.Name() == templateName {
			return true
		}
	}
	return false
}
