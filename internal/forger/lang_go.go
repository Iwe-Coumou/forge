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

func (goLang) Keys() Keys {
	return Keys{
		// module_path is per-project, so it is flag-only; base_module is the
		// shared prefix it is derived from, so it is config-only.
		Flag:   []string{"module_path", "go_version"},
		Config: []string{"base_module", "go_version"},
	}
}

func (goLang) Context(p *Project, cfg *config.Config) (any, error) {
	mod := p.Overrides["module_path"]
	if mod == "" {
		mod = cfg.ModulePathFor(p.Name)
	}

	goVersion := setting(p, cfg, "go_version", "1.22")

	return GoContext{
		Common:     commonContext(p, cfg),
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
