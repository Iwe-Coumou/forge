package forger

import (
	"fmt"
	"strings"
	"time"

	"github.com/Iwe-Coumou/forge/v2/internal/config"
)

// Common holds the render-context fields every language provides.
type Common struct {
	Name    string
	Author  string
	Email   string
	License string
	Year    int
}

func commonContext(p *Project, cfg *config.Config) Common {
	return Common{
		Name:    p.Name,
		Author:  cfg.Author,
		Email:   cfg.Email,
		License: cfg.License,
		Year:    time.Now().Year(),
	}
}

// Keys declares the tunable values a language exposes, and is the single
// source of truth for what `--set` and the config file accept.
//
// The two lists overlap but are not identical: module_path is per-project so
// it is flag-only, while base_module is a shared prefix so it is config-only.
type Keys struct {
	Flag   []string // accepted by --set
	Config []string // accepted under languages.<name> in the config file
}

type Language interface {
	Name() string
	Keys() Keys
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

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// allowedList renders a key list for an error message.
func allowedList(keys []string) string {
	if len(keys) == 0 {
		return "it accepts none"
	}
	return "allowed: " + strings.Join(keys, ", ")
}

// checkOverrides rejects --set keys the language does not accept. Forge calls
// this once per scaffold, so no language has to remember to.
func checkOverrides(l Language, p *Project) error {
	allowed := l.Keys().Flag

	for k := range p.Overrides {
		if !contains(allowed, k) {
			return fmt.Errorf("unknown override %q for language %q (%s)", k, l.Name(), allowedList(allowed))
		}
	}
	return nil
}

// CheckConfig rejects config keys that a registered language does not accept,
// so a typo in the config file fails as loudly as one on the command line.
// Sections for languages this build doesn't know about are ignored, so a
// config written for a newer Forge still works.
func CheckConfig(cfg *config.Config) error {
	for langName, settings := range cfg.Languages {
		l, err := lookupLanguage(langName)
		if err != nil {
			continue
		}

		allowed := l.Keys().Config
		for key := range settings {
			if !contains(allowed, key) {
				return fmt.Errorf("unknown config key %q under languages.%s (%s)", key, langName, allowedList(allowed))
			}
		}
	}
	return nil
}

// setting resolves a language value: --set flag first, then the config file,
// then the language's built-in default.
func setting(p *Project, cfg *config.Config, key, fallback string) string {
	if v := p.Overrides[key]; v != "" {
		return v
	}
	if v := cfg.LanguageSetting(p.Language, key); v != "" {
		return v
	}
	return fallback
}
