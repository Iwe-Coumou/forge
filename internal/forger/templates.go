package forger

import (
	"fmt"
	"io/fs"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

type TemplateInfo struct {
	Language string
	Name     string
	Short    string
	Long     string

	// NotImplemented is non-empty when the template's language is registered
	// but not yet usable, and holds the reason why.
	NotImplemented string
}

// ID returns the qualified "language/name" identifier used on the CLI.
func (t TemplateInfo) ID() string { return t.Language + "/" + t.Name }

type templateMeta struct {
	Language string `yaml:"language"`
	Short    string `yaml:"short"`
	Long     string `yaml:"long"`
}

// ListTemplates returns metadata for every available template.
func ListTemplates() ([]TemplateInfo, error) {
	langs, err := templateFS.ReadDir("templates")
	if err != nil {
		return nil, err
	}

	var templates []TemplateInfo
	for _, lang := range langs {
		if !lang.IsDir() {
			continue
		}

		entries, err := templateFS.ReadDir("templates/" + lang.Name())
		if err != nil {
			return nil, err
		}

		for _, e := range entries {
			if !e.IsDir() {
				continue
			}

			info := TemplateInfo{Language: lang.Name(), Name: e.Name()}
			if l, err := lookupLanguage(lang.Name()); err == nil {
				info.NotImplemented = notImplementedReason(l)
			}

			data, err := templateFS.ReadFile(path.Join("templates", lang.Name(), e.Name(), "template.yaml"))
			if err != nil {
				// no metadata file — still list it, just without descriptions
				templates = append(templates, info)
				continue
			}

			var meta templateMeta
			if err := yaml.Unmarshal(data, &meta); err != nil {
				return nil, err
			}
			info.Short = meta.Short
			info.Long = meta.Long

			templates = append(templates, info)
		}
	}

	return templates, nil
}

// ParseTemplateID splits a "language/template" argument into its parts
// and verifies the template exists.
func ParseTemplateID(arg string) (string, string, error) {
	arg = strings.ReplaceAll(arg, "\\", "/")

	lang, name, found := strings.Cut(arg, "/")
	if !found || lang == "" || name == "" || strings.Contains(name, "/") {
		return "", "", fmt.Errorf("invalid template %q, want language/template (e.g. go/cli_cobra)", arg)
	}

	if !templateExists(lang, name) {
		return "", "", fmt.Errorf("unknown template %q", arg)
	}

	return lang, name, nil
}

func templateExists(lang, name string) bool {
	_, err := fs.Stat(templateFS, "templates/"+lang+"/"+name+"/template.yaml")
	return err == nil
}
