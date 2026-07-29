package forger

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/fatih/color"
)

//go:embed templates
var templateFS embed.FS

func Forge(p *Project, verbose bool) error {
	if err := p.validate(); err != nil {
		return fmt.Errorf("invalid project: %w", err)
	}

	if verbose {
		color.Cyan("forging %q from %q into %s\n", p.Name, p.Template, p.OutputDir)
	}

	templateRoot := "templates/" + p.Template

	return fs.WalkDir(templateFS, templateRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if d.Name() == "template.yaml" {
			return nil
		}

		relPath, err := filepath.Rel(templateRoot, path)
		if err != nil {
			return err
		}
		relPath = strings.TrimSuffix(relPath, ".tmpl")
		destPath := filepath.Join(p.OutputDir, relPath)

		if verbose {
			fmt.Printf("rendering %s -> %s\n", path, destPath)
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return err
		}

		tmpl, err := template.ParseFS(templateFS, path)
		if err != nil {
			return err
		}

		out, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer out.Close()

		return tmpl.Execute(out, p)
	})
}
