package config

import (
	"testing"
)

// withTempConfigDir redirects os.UserConfigDir() to a temp directory for
// the duration of the test, regardless of which OS it runs on.
func withTempConfigDir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("AppData", dir)         // windows
	t.Setenv("XDG_CONFIG_HOME", dir) // linux
	t.Setenv("HOME", dir)            // darwin fallback
}

func TestDefault(t *testing.T) {
	cfg := Default()
	if got := cfg.LanguageSetting("go", "base_module"); got != "local" {
		t.Errorf("Default() go base_module = %q, want %q", got, "local")
	}
	if cfg.BaseModule != "" {
		t.Errorf("Default().BaseModule = %q, want empty — new configs must not write the deprecated key", cfg.BaseModule)
	}
}

func TestLanguageSetting(t *testing.T) {
	populated := Config{
		Languages: map[string]map[string]string{
			"go": {"go_version": "1.23"},
		},
	}

	tests := []struct {
		name string
		cfg  Config
		lang string
		key  string
		want string
	}{
		{name: "no languages block", cfg: Config{}, lang: "go", key: "go_version", want: ""},
		{name: "unknown language", cfg: populated, lang: "rust", key: "go_version", want: ""},
		{name: "unknown key", cfg: populated, lang: "go", key: "nope", want: ""},
		{name: "configured", cfg: populated, lang: "go", key: "go_version", want: "1.23"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.LanguageSetting(tt.lang, tt.key); got != tt.want {
				t.Errorf("LanguageSetting(%q, %q) = %q, want %q", tt.lang, tt.key, got, tt.want)
			}
		})
	}
}

func TestModulePathFor(t *testing.T) {
	nested := func(base string) map[string]map[string]string {
		return map[string]map[string]string{"go": {"base_module": base}}
	}

	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{
			name: "nothing configured",
			cfg:  Config{},
			want: "myproj",
		},
		{
			// Pins the deprecation contract: configs written before the
			// languages block must keep resolving. Don't delete BaseModule
			// without a migration.
			name: "legacy top-level base_module",
			cfg:  Config{BaseModule: "github.com/old"},
			want: "github.com/old/myproj",
		},
		{
			name: "languages.go.base_module",
			cfg:  Config{Languages: nested("github.com/new")},
			want: "github.com/new/myproj",
		},
		{
			name: "nested wins over legacy",
			cfg:  Config{BaseModule: "github.com/old", Languages: nested("github.com/new")},
			want: "github.com/new/myproj",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.ModulePathFor("myproj"); got != tt.want {
				t.Errorf("ModulePathFor() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSaveLoadRoundTrip_Languages(t *testing.T) {
	withTempConfigDir(t)

	want := Config{
		Languages: map[string]map[string]string{
			"go":     {"base_module": "github.com/someone", "go_version": "1.23"},
			"python": {"min_python": "3.12"},
		},
	}
	if err := Save(&want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	for _, tc := range []struct{ lang, key, want string }{
		{"go", "base_module", "github.com/someone"},
		{"go", "go_version", "1.23"},
		{"python", "min_python", "3.12"},
	} {
		if got := got.LanguageSetting(tc.lang, tc.key); got != tc.want {
			t.Errorf("LanguageSetting(%q, %q) = %q, want %q", tc.lang, tc.key, got, tc.want)
		}
	}

	// omitempty means a config with no legacy key doesn't grow one.
	if got.BaseModule != "" {
		t.Errorf("BaseModule = %q, want empty after a round trip that never set it", got.BaseModule)
	}
}

func TestExists_NoConfig(t *testing.T) {
	withTempConfigDir(t)

	exists, err := Exists()
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if exists {
		t.Error("Exists() = true, want false before any config is saved")
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	withTempConfigDir(t)

	want := Config{BaseModule: "github.com/someone"}
	if err := Save(&want); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	exists, err := Exists()
	if err != nil {
		t.Fatalf("Exists() error = %v", err)
	}
	if !exists {
		t.Error("Exists() = false after Save(), want true")
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if got.BaseModule != want.BaseModule {
		t.Errorf("Load().BaseModule = %q, want %q", got.BaseModule, want.BaseModule)
	}
}

func TestLoad_MissingFileReturnsZeroValue(t *testing.T) {
	withTempConfigDir(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.BaseModule != "" {
		t.Errorf("Load().BaseModule = %q, want empty when no config exists", cfg.BaseModule)
	}
}
