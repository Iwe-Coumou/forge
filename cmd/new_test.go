package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

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
