package forger

import (
	"gopkg.in/yaml.v3"
)

type TemplateInfo struct {
	Name  string
	Short string
	Long  string
}

type templateMeta struct {
	Short string `yaml:"short"`
	Long  string `yaml:"long"`
}

// ListTemplates returns metadata for every available template.
func ListTemplates() ([]TemplateInfo, error) {
	entries, err := templateFS.ReadDir("templates")
	if err != nil {
		return nil, err
	}

	var templates []TemplateInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		data, err := templateFS.ReadFile("templates/" + e.Name() + "/template.yaml")
		if err != nil {
			// no metadata file — still list it, just without descriptions
			templates = append(templates, TemplateInfo{Name: e.Name()})
			continue
		}

		var meta templateMeta
		if err := yaml.Unmarshal(data, &meta); err != nil {
			return nil, err
		}

		templates = append(templates, TemplateInfo{
			Name:  e.Name(),
			Short: meta.Short,
			Long:  meta.Long,
		})
	}

	return templates, nil
}
