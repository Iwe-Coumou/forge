package forger

import (
	"fmt"
	"os"
)

type Project struct {
	Name      string
	OutputDir string
	Language  string
	Template  string
	Overrides map[string]string // raw user input, e.g. --set module_path=...
}

func (p *Project) validate() error {
	if p.Name == "" {
		return fmt.Errorf("project name is required")
	}

	if p.OutputDir == "" {
		return fmt.Errorf("project output directory is required")
	}

	if p.Language == "" {
		return fmt.Errorf("project language is required")
	}

	if p.Template == "" {
		return fmt.Errorf("project template is required")
	}

	if !templateExists(p.Language, p.Template) {
		return fmt.Errorf("unknown template %q for language %q", p.Template, p.Language)
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
