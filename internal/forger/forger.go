package forger

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/Iwe-Coumou/forge/v2/internal/config"
	"github.com/fatih/color"
)

//go:embed all:templates
var templateFS embed.FS

func Forge(p *Project, cfg *config.Config, verbose bool) error {
	if err := p.validate(); err != nil {
		return fmt.Errorf("invalid project: %w", err)
	}

	if verbose {
		color.Cyan("forging %q from %q into %s\n", p.Name, p.Language+"/"+p.Template, p.OutputDir)
	}

	lang, err := lookupLanguage(p.Language)
	if err != nil {
		return err
	}
	if reason := notImplementedReason(lang); reason != "" {
		return fmt.Errorf("%s support is not implemented yet: %s", lang.Name(), reason)
	}

	// Validated centrally so a language can't forget to, and so a typo in
	// either place fails before anything is written.
	if err := checkOverrides(lang, p); err != nil {
		return err
	}
	if err := CheckConfig(cfg); err != nil {
		return err
	}

	ctx, err := lang.Context(p, cfg)
	if err != nil {
		return fmt.Errorf("getting context: %w", err)
	}

	return walkTemplate(p.Language, p.Template, func(fsPath, relPath string) error {
		destPath := filepath.Join(p.OutputDir, filepath.FromSlash(relPath))

		if verbose {
			fmt.Printf("rendering %s -> %s\n", fsPath, destPath)
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			return err
		}

		tmpl, err := template.ParseFS(templateFS, fsPath)
		if err != nil {
			return err
		}

		out, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer out.Close()

		return tmpl.Execute(out, ctx)
	})
}
