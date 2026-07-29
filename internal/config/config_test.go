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
	if cfg.BaseModule != "local" {
		t.Errorf("Default().BaseModule = %q, want %q", cfg.BaseModule, "local")
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
