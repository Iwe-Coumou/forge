package forger

import (
	"fmt"
	"strings"

	"github.com/Iwe-Coumou/forge/internal/config"
)

// Common holds the render-context fields every language provides.
type Common struct {
	Name string
}

type Language interface {
	Name() string
	Context(p *Project, cfg *config.Config) (any, error)
	PostProcess(dir string, verbose bool) error
	VerifyCmd() []string
}

// Unimplemented is an optional interface a Language can implement to mark
// itself as registered but not yet usable. Forge refuses to scaffold such a
// language, and `forge list` flags its templates. Delete the method to
// enable the language.
type Unimplemented interface {
	NotImplementedReason() string
}

// notImplementedReason returns why l is unusable, or "" when it is ready.
func notImplementedReason(l Language) string {
	if u, ok := l.(Unimplemented); ok {
		return u.NotImplementedReason()
	}
	return ""
}

var languages = map[string]Language{}

func RegisterLanguage(l Language) {
	name := l.Name()
	if name == "" {
		panic("forger: language with empty name")
	}
	if _, exists := languages[name]; exists {
		panic("forger: duplicate language " + name)
	}
	languages[name] = l
}

func lookupLanguage(name string) (Language, error) {
	l, ok := languages[name]
	if !ok {
		return nil, fmt.Errorf("unknown language %q", name)
	}
	return l, nil
}

func checkOverrides(p *Project, allowed ...string) error {
	for k := range p.Overrides {
		found := false
		for _, a := range allowed {
			if k == a {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("unknown override %q for language %q (allowed %s)", k, p.Language, strings.Join(allowed, ", "))
		}
	}
	return nil
}
