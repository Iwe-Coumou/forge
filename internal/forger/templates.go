package forger

import (
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"sort"
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

			info, err := readTemplateInfo(lang.Name(), e.Name())
			if err != nil {
				return nil, err
			}
			templates = append(templates, info)
		}
	}

	return templates, nil
}

// readTemplateInfo loads one template's metadata. A template with no
// template.yaml is still valid, just undescribed.
func readTemplateInfo(lang, name string) (TemplateInfo, error) {
	info := TemplateInfo{Language: lang, Name: name}

	if l, err := lookupLanguage(lang); err == nil {
		info.NotImplemented = notImplementedReason(l)
	}

	data, err := templateFS.ReadFile(path.Join("templates", lang, name, "template.yaml"))
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

// walkTemplate calls fn for every renderable file in a template. fsPath is
// the file inside templateFS; relPath is its slash-separated destination
// relative to the project root, with any .tmpl suffix stripped.
//
// Forge and InspectTemplate both go through this, so what inspect lists is
// exactly what forge writes.
func walkTemplate(lang, name string, fn func(fsPath, relPath string) error) error {
	root := "templates/" + lang + "/" + name

	return fs.WalkDir(templateFS, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() == "template.yaml" {
			return nil
		}

		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}

		return fn(p, strings.TrimSuffix(filepath.ToSlash(rel), ".tmpl"))
	})
}

// InspectTemplate returns the details of a single template.
func InspectTemplate(lang, name string) (*TemplateDetail, error) {
	if !templateExists(lang, name) {
		return nil, fmt.Errorf("unknown template %q", lang+"/"+name)
	}

	info, err := readTemplateInfo(lang, name)
	if err != nil {
		return nil, err
	}

	detail := &TemplateDetail{TemplateInfo: info}

	// A template whose language is not registered is a bug, but inspect is
	// the command you reach for while diagnosing it — so degrade rather than
	// refuse, and leave VerifyCmd empty.
	if l, err := lookupLanguage(lang); err == nil {
		detail.VerifyCmd = l.VerifyCmd()
	}

	if err := walkTemplate(lang, name, func(_, relPath string) error {
		detail.Files = append(detail.Files, relPath)
		return nil
	}); err != nil {
		return nil, err
	}
	sort.Strings(detail.Files)

	return detail, nil
}
