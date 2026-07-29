package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Iwe-Coumou/forge/internal/config"
)

// isolateConfigDir redirects the config package's os.UserConfigDir() lookup
// to a temp directory so tests don't read or write the real user config.
func isolateConfigDir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("AppData", dir)         // windows
	t.Setenv("XDG_CONFIG_HOME", dir) // linux
	t.Setenv("HOME", dir)            // darwin fallback
}

func TestResolveOutputDir_WithExplicitPath(t *testing.T) {
	old := projectPath
	defer func() { projectPath = old }()
	projectPath = filepath.Join("some", "base")

	got, err := resolveOutputDir("myproj")
	if err != nil {
		t.Fatalf("resolveOutputDir() error = %v", err)
	}
	want := filepath.Join("some", "base", "myproj")
	if got != want {
		t.Errorf("resolveOutputDir() = %q, want %q", got, want)
	}
}

func TestResolveOutputDir_DefaultsToCwd(t *testing.T) {
	old := projectPath
	defer func() { projectPath = old }()
	projectPath = ""

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	got, err := resolveOutputDir("myproj")
	if err != nil {
		t.Fatalf("resolveOutputDir() error = %v", err)
	}
	want := filepath.Join(cwd, "myproj")
	if got != want {
		t.Errorf("resolveOutputDir() = %q, want %q", got, want)
	}
}

func TestResolveModulePath_NoBaseModuleConfigured(t *testing.T) {
	isolateConfigDir(t)

	got, err := resolveModulePath("myproj")
	if err != nil {
		t.Fatalf("resolveModulePath() error = %v", err)
	}
	if got != "myproj" {
		t.Errorf("resolveModulePath() = %q, want %q", got, "myproj")
	}
}

func TestResolveModulePath_WithBaseModuleConfigured(t *testing.T) {
	isolateConfigDir(t)

	if err := config.Save(&config.Config{BaseModule: "github.com/someone"}); err != nil {
		t.Fatalf("config.Save() error = %v", err)
	}

	got, err := resolveModulePath("myproj")
	if err != nil {
		t.Fatalf("resolveModulePath() error = %v", err)
	}
	want := "github.com/someone/myproj"
	if got != want {
		t.Errorf("resolveModulePath() = %q, want %q", got, want)
	}
}
