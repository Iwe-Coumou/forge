package forger

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Iwe-Coumou/forge/v2/internal/config"
	"go.yaml.in/yaml/v3"
)

type TemplateInfo struct {
	Language string
	Name     string
	Short    string
	Long     string

	// NotImplemented is non-empty when the template's language is registered
	// but not yet usable, and holds the reason why.
	NotImplemented string

	// Source is where this template was found: "embedded", or the user
	// template directory it came from.
	Source string
}

// ID returns the qualified "language/name" identifier used on the CLI.
func (t TemplateInfo) ID() string { return t.Language + "/" + t.Name }

type templateMeta struct {
	Language string `yaml:"language"`
	Short    string `yaml:"short"`
	Long     string `yaml:"long"`
}

// TemplateDetail is everything `forge inspect` needs about one template: its
// metadata, the files it generates, and what its language supports.
type TemplateDetail struct {
	TemplateInfo

	// Files are the paths this template generates, slash-separated and
	// relative to the project root, in lexical order.
	Files []string

	// VerifyCmd is the command that proves a generated project is valid.
	// Empty when the template's language is not registered.
	VerifyCmd []string

	// Keys are the values this template's language exposes to --set and to
	// the config file.
	Keys Keys
}

type templateSource struct {
	fsys  fs.FS
	label string // "embedded", or the directory path, for diagnostics
}

// templateSources returns the sources to search, embedded first.
func templateSources(cfg *config.Config) []templateSource {
	sources := []templateSource{{fsys: embeddedTemplates(), label: "embedded"}}

	dir, err := cfg.TemplatesDir()
	if err != nil || dir == "" {
		return sources
	}
	if info, statErr := os.Stat(dir); statErr != nil || !info.IsDir() {
		return sources // absent is normal, not an error
	}

	return append(sources, templateSource{fsys: os.DirFS(dir), label: dir})
}

// ListTemplates returns metadata for every available template.
func ListTemplates(cfg *config.Config) ([]TemplateInfo, error) {
	var templates []TemplateInfo
	seen := map[string]string{}

	for _, src := range templateSources(cfg) {
		langs, err := fs.ReadDir(src.fsys, ".")
		if err != nil {
			return nil, err
		}

		for _, lang := range langs {
			if !lang.IsDir() {
				continue
			}

			entries, err := fs.ReadDir(src.fsys, lang.Name())
			if err != nil {
				return nil, err
			}

			for _, e := range entries {
				if !e.IsDir() {
					continue
				}

				info, err := readTemplateInfo(src.fsys, lang.Name(), e.Name())
				if err != nil {
					return nil, err
				}
				info.Source = src.label

				if prev, dup := seen[info.ID()]; dup {
					return nil, fmt.Errorf("template %s defined in both %s and %s", info.ID(), prev, src.label)
				}
				seen[info.ID()] = src.label

				templates = append(templates, info)
			}
		}
	}

	return templates, nil
}

// readTemplateInfo loads one template's metadata. A template with no
// template.yaml is still valid, just undescribed.
func readTemplateInfo(fsys fs.FS, lang, name string) (TemplateInfo, error) {
	info := TemplateInfo{Language: lang, Name: name}

	if l, err := lookupLanguage(lang); err == nil {
		info.NotImplemented = notImplementedReason(l)
	}

	data, err := fs.ReadFile(fsys, path.Join(lang, name, "template.yaml"))
	if err != nil {
		return info, nil
	}

	var meta templateMeta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return info, err
	}
	info.Short = meta.Short
	info.Long = meta.Long

	return info, nil
}

// ParseTemplateID splits a "language/template" argument into its parts
// and verifies the template exists.
func ParseTemplateID(cfg *config.Config, arg string) (string, string, error) {
	arg = strings.ReplaceAll(arg, "\\", "/")

	lang, name, found := strings.Cut(arg, "/")
	if !found || lang == "" || name == "" || strings.Contains(name, "/") {
		return "", "", fmt.Errorf("invalid template %q, want language/template (e.g. go/cli_cobra)", arg)
	}

	if _, err := findTemplate(cfg, lang, name); err != nil {
		return "", "", err
	}

	return lang, name, nil
}

func templateExists(fsys fs.FS, lang, name string) bool {
	_, err := fs.Stat(fsys, path.Join(lang, name, "template.yaml"))
	return err == nil
}

// walkTemplate calls fn for every renderable file in a template. fsPath is
// the file inside fsys; relPath is its slash-separated destination
// relative to the project root, with any .tmpl suffix stripped.
//
// Forge and InspectTemplate both go through this, so what inspect lists is
// exactly what forge writes.
func walkTemplate(fsys fs.FS, lang, name string, fn func(fsPath, relPath string) error) error {
	root := path.Join(lang, name)

	return fs.WalkDir(fsys, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() == "template.yaml" {
			return nil
		}

		rel := strings.TrimPrefix(p, root+"/")
		dest := strings.TrimSuffix(rel, ".tmpl")

		// Templates can come from outside the binary, so a path must never
		// escape the project directory or name a reserved device.
		if !filepath.IsLocal(filepath.FromSlash(dest)) {
			return fmt.Errorf("template file %q would escape the project directory", p)
		}
		return fn(p, dest)
	})
}

// InspectTemplate returns the details of a single template.
func InspectTemplate(cfg *config.Config, lang, name string) (*TemplateDetail, error) {
	src, err := findTemplate(cfg, lang, name)
	if err != nil {
		return nil, err
	}

	info, err := readTemplateInfo(src.fsys, lang, name)
	if err != nil {
		return nil, err
	}
	info.Source = src.label

	detail := &TemplateDetail{TemplateInfo: info}

	// A template whose language is not registered is a bug, but inspect is
	// the command you reach for while diagnosing it — so degrade rather than
	// refuse, and leave VerifyCmd empty.
	if l, err := lookupLanguage(lang); err == nil {
		detail.VerifyCmd = l.VerifyCmd()
		detail.Keys = l.Keys()
	}

	if err := walkTemplate(src.fsys, lang, name, func(_, relPath string) error {
		detail.Files = append(detail.Files, relPath)
		return nil
	}); err != nil {
		return nil, err
	}
	sort.Strings(detail.Files)

	return detail, nil
}

// findTemplate returns the source holding a template.
func findTemplate(cfg *config.Config, lang, name string) (templateSource, error) {
	for _, src := range templateSources(cfg) {
		if templateExists(src.fsys, lang, name) {
			return src, nil
		}
	}
	return templateSource{}, fmt.Errorf("unknown template %q", lang+"/"+name)
}
