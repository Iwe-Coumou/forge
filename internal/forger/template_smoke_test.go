package forger

import (
	"os/exec"
	"path/filepath"
	"testing"
)

// TestTemplatesCompile scaffolds every registered template into a temp
// directory, tidies and formats it exactly as `forge new` does, then
// builds it — catching template files that don't produce valid Go.
func TestTemplatesCompile(t *testing.T) {
	templates, err := ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}
	if len(templates) == 0 {
		t.Fatal("no templates registered")
	}

	for _, tmpl := range templates {
		t.Run(tmpl.Name, func(t *testing.T) {
			outputDir := filepath.Join(t.TempDir(), "smoketest")

			p := &Project{
				Name:       "smoketest",
				ModulePath: "example.com/smoketest",
				OutputDir:  outputDir,
				Template:   tmpl.Name,
			}

			if err := Forge(p, false); err != nil {
				t.Fatalf("Forge() error = %v", err)
			}

			if err := PostProcess(outputDir, &PostProcessOptions{GitInit: false}, false); err != nil {
				t.Fatalf("PostProcess() error = %v", err)
			}

			cmd := exec.Command("go", "build", "./...")
			cmd.Dir = outputDir
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("go build ./... failed: %v\n%s", err, out)
			}
		})
	}
}
