package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

type Config struct {
	// Deprecated: use languages.go.base_module. Still read so configs written
	// before v0.x keep working; new configs no longer write this key.
	BaseModule string `yaml:"base_module,omitempty"`

	Author  string `yaml:"author,omitempty"`
	Email   string `yaml:"email,omitempty"`
	License string `yaml:"license,omitempty"`
	GitInit bool   `yaml:"git_init,omitempty"`

	TemplatesDirectory string `yaml:"templates_dir,omitempty"`

	Languages map[string]map[string]string `yaml:"languages"`
}

func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return &Config{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return &Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// Exists reports whether a config file has already been created.
func Exists() (bool, error) {
	path, err := configPath()
	if err != nil {
		return false, err
	}

	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func Default() Config {
	return Config{
		Languages: map[string]map[string]string{
			"go": {"base_module": "local"},
		},
	}
}

func configPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, "forge", "config.yaml"), nil
}

// LanguageSettings returns a configured value for a language, or "" when it
// isn't set.
func (c *Config) LanguageSetting(lang, key string) string {
	return c.Languages[lang][key]
}

func (c *Config) ModulePathFor(projectName string) string {
	base := c.LanguageSetting("go", "base_module")
	if base == "" {
		base = c.BaseModule
	}
	if base == "" {
		return projectName
	}
	return base + "/" + projectName
}

// TemplatesDir returns where user templates live: the configured path if set,
// otherwise a "templates" directory alongside the config file.
func (c *Config) TemplatesDir() (string, error) {
	if c.TemplatesDirectory != "" {
		return expandHome(c.TemplatesDirectory)
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "forge", "templates"), nil
}

// expandHome replaces a leading ~ with the user's home directory. Anything
// else is returned unchanged. A bare ~ elsewhere in the path is a literal
// directory name, not a shorthand.
func expandHome(p string) (string, error) {
	if p != "~" && !strings.HasPrefix(p, "~/") && !strings.HasPrefix(p, "~\\") {
		return p, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	if p == "~" {
		return home, nil
	}
	return filepath.Join(home, p[2:]), nil
}
