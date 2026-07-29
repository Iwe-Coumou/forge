package forger

import (
	"fmt"

	"github.com/Iwe-Coumou/forge/internal/config"
)

// GoContext is the data Go templates are rendered against.
type GoContext struct {
	Common
	ModulePath string
	GoVersion  string
}

type goLang struct{}

func init() {
	RegisterLanguage(goLang{})
}

func (goLang) Name() string { return "go" }

func (goLang) Context(p *Project, cfg *config.Config) (any, error) {
	if err := checkOverrides(p, "module_path", "go_version"); err != nil {
		return nil, err
	}

	mod := p.Overrides["module_path"]
	if mod == "" {
		mod = cfg.ModulePathFor(p.Name)
	}

	goVersion := p.Overrides["go_version"]
	if goVersion == "" {
		goVersion = "1.22"
	}

	return GoContext{
		Common:     Common{Name: p.Name},
		ModulePath: mod,
		GoVersion:  goVersion,
	}, nil
}

func (goLang) PostProcess(dir string, verbose bool) error {
	if err := runIn(dir, verbose, "go", "mod", "tidy"); err != nil {
		return fmt.Errorf("go mod tidy: %w", err)
	}

	if err := runIn(dir, verbose, "gofmt", "-w", "."); err != nil {
		return fmt.Errorf("gofmt: %w", err)
	}
	return nil
}

func (goLang) VerifyCmd() []string { return []string{"go", "build", "./..."} }
