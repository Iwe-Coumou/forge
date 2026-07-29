package forger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestForgeRendersTemplateFiles(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "myproj")

	p := &Project{
		Name:       "myproj",
		ModulePath: "example.com/myproj",
		OutputDir:  outputDir,
		Template:   "cli_cobra",
	}

	if err := Forge(p, false); err != nil {
		t.Fatalf("Forge() error = %v", err)
	}

	var got []string
	err := filepath.WalkDir(outputDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			rel, err := filepath.Rel(outputDir, path)
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
