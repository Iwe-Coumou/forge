package forger

import (
	"github.com/fatih/color"
)

type Project struct {
	Name       string
	ModulePath string
	OutputDir  string
	Template   string
}

func Forge(p *Project, verbose bool) error {
	if verbose {
		color.Cyan("forging %q from %q into %s\n", p.Name, p.Template, p.OutputDir)
	}
	return nil
}
