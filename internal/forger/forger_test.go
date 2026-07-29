package forger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Iwe-Coumou/forge/internal/config"
)

// renderedFiles returns every file under dir, as slash-separated paths
// relative to dir.
func renderedFiles(t *testing.T, dir string) []string {
	t.Helper()

	var got []string
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			got = append(got, filepath.ToSlash(rel))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking output dir: %v", err)
	}
	return got
}

func TestForgeRendersTemplateFiles(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "myproj")

	p := &Project{
		Name:      "myproj",
		OutputDir: outputDir,
		Language:  "go",
		Template:  "cli_cobra",
	}
	cfg := &config.Config{BaseModule: "example.com"}

	if err := Forge(p, cfg, false); err != nil {
		t.Fatalf("Forge() error = %v", err)
	}

	got := renderedFiles(t, outputDir)

	want := []string{"cmd/example.go", "cmd/root.go", "go.mod", "main.go"}
	if len(got) != len(want) {
		t.Fatalf("rendered files = %v, want %v", got, want)
	}
	for _, w := range want {
		found := false
		for _, g := range got {
			if g == w {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing rendered file %q, got %v", w, got)
		}
	}

	for _, f := range got {
		if strings.HasSuffix(f, ".tmpl") {
			t.Errorf("rendered file %q still has .tmpl suffix", f)
		}
		if filepath.Base(f) == "template.yaml" {
			t.Errorf("template.yaml metadata leaked into rendered output")
		}
	}

	mainGo, err := os.ReadFile(filepath.Join(outputDir, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(mainGo), `"example.com/myproj/cmd"`) {
		t.Errorf("main.go = %s, want import of %q", mainGo, "example.com/myproj/cmd")
	}

	goMod, err := os.ReadFile(filepath.Join(outputDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(goMod), "module example.com/myproj") {
		t.Errorf("go.mod = %s, want module declaration for %q", goMod, "example.com/myproj")
	}

	rootGo, err := os.ReadFile(filepath.Join(outputDir, "cmd", "root.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(rootGo), `"myproj"`) {
		t.Errorf("cmd/root.go = %s, want project name %q", rootGo, "myproj")
	}
}

// TestForgeRejectsUnimplementedLanguage covers languages marked via the
// Unimplemented interface: Forge must refuse before writing any files.
func TestForgeRejectsUnimplementedLanguage(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "my_proj")

	p := &Project{
		Name:      "my_proj",
		OutputDir: outputDir,
		Language:  "python",
		Template:  "cli",
	}

	err := Forge(p, &config.Config{}, false)
	if err == nil {
		t.Fatal("Forge() = nil, want error for an unimplemented language")
	}
	if !strings.Contains(err.Error(), "not implemented yet") {
		t.Errorf("Forge() error = %v, want it to say the language is not implemented", err)
	}

	if _, statErr := os.Stat(outputDir); !os.IsNotExist(statErr) {
		t.Errorf("Forge() created %s, want no output for a refused language", outputDir)
	}
}

// TestPythonContextRenders exercises the Python render path directly, so the
// second language keeps coverage while it is still marked unimplemented.
func TestPythonContextRenders(t *testing.T) {
	lang, err := lookupLanguage("python")
	if err != nil {
		t.Fatalf("lookupLanguage() error = %v", err)
	}

	ctx, err := lang.Context(&Project{Name: "my_proj"}, &config.Config{})
	if err != nil {
		t.Fatalf("Context() error = %v", err)
	}

	got, ok := ctx.(PythonContext)
	if !ok {
		t.Fatalf("Context() = %T, want PythonContext", ctx)
	}
	if got.Name != "my_proj" {
		t.Errorf("Name = %q, want %q", got.Name, "my_proj")
	}
	if got.DistName != "my-proj" {
		t.Errorf("DistName = %q, want %q", got.DistName, "my-proj")
	}
	if got.ImportName != "my_proj" {
		t.Errorf("ImportName = %q, want %q", got.ImportName, "my_proj")
	}
}

func TestForgeRejectsUnknownOverride(t *testing.T) {
	p := &Project{
		Name:      "myproj",
		OutputDir: filepath.Join(t.TempDir(), "myproj"),
		Language:  "go",
		Template:  "cli_cobra",
		Overrides: map[string]string{"modulepath": "example.com/typo"},
	}

	err := Forge(p, &config.Config{}, false)
	if err == nil {
		t.Fatal("Forge() = nil, want error for unknown override key")
	}
	if !strings.Contains(err.Error(), "modulepath") {
		t.Errorf("Forge() error = %v, want it to name the offending key", err)
	}
}

func TestForgeAppliesOverrides(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "myproj")

	p := &Project{
		Name:      "myproj",
		OutputDir: outputDir,
		Language:  "go",
		Template:  "cli_cobra",
		Overrides: map[string]string{"module_path": "example.org/override"},
	}

	// BaseModule would normally win; the override must take precedence.
	cfg := &config.Config{BaseModule: "example.com"}
	if err := Forge(p, cfg, false); err != nil {
		t.Fatalf("Forge() error = %v", err)
	}

	goMod, err := os.ReadFile(filepath.Join(outputDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(goMod), "module example.org/override") {
		t.Errorf("go.mod = %s, want overridden module path", goMod)
	}
}
