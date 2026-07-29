package forger

import (
	"strings"

	"github.com/Iwe-Coumou/forge/internal/config"
)

// PythonContext is the data Python templates are rendered against.
type PythonContext struct {
	Common
	DistName   string // my-app
	ImportName string // my_app
	MinPython  string
}

type pythonLang struct{}

func init() {
	RegisterLanguage(pythonLang{})
}

func (pythonLang) Name() string { return "python" }

// NotImplementedReason marks Python as a work in progress. Delete this method
// once the templates and post-processing are real, and Python becomes usable
// with no other change.
func (pythonLang) NotImplementedReason() string {
	return "the templates are stubs and no formatting or packaging is set up"
}

func (pythonLang) Context(p *Project, cfg *config.Config) (any, error) {
	if err := checkOverrides(p, "min_python"); err != nil {
		return nil, err
	}

	minPython := p.Overrides["min_python"]
	if minPython == "" {
		minPython = "3.11"
	}

	return PythonContext{
		Common:     Common{Name: p.Name},
		DistName:   strings.ReplaceAll(p.Name, "_", "-"),
		ImportName: strings.ReplaceAll(p.Name, "-", "_"),
		MinPython:  minPython,
	}, nil
}

func (pythonLang) PostProcess(dir string, verbose bool) error {
	return nil
}

func (pythonLang) VerifyCmd() []string { return []string{"python", "-m", "compileall", "."} }
