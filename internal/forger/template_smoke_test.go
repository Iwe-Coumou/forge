package forger

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/Iwe-Coumou/forge/internal/config"
)

// TestTemplatesCompile scaffolds every registered template into a temp
// directory, post-processes it exactly as `forge new` does, then runs the
// language's verify command — catching template files that don't produce
// valid code. Templates whose toolchain isn't installed are skipped.
func TestTemplatesCompile(t *testing.T) {
	templates, err := ListTemplates()
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}
	if len(templates) == 0 {
		t.Fatal("no templates registered")
	}

	cfg := &config.Config{BaseModule: "example.com"}

	for _, tmpl := range templates {
		t.Run(tmpl.ID(), func(t *testing.T) {
			lang, err := lookupLanguage(tmpl.Language)
			if err != nil {
				t.Fatalf("template %s: %v", tmpl.ID(), err)
			}

			if reason := notImplementedReason(lang); reason != "" {
				t.Skipf("%s is not implemented yet: %s", tmpl.Language, reason)
			}

			verify := lang.VerifyCmd()
			if len(verify) == 0 {
				t.Fatalf("language %q has an empty VerifyCmd", tmpl.Language)
			}
			if _, err := exec.LookPath(verify[0]); err != nil {
				t.Skipf("%s is not installed, skipping", verify[0])
			}

			outputDir := filepath.Join(t.TempDir(), "smoketest")

			p := &Project{
				Name:      "smoketest",
				OutputDir: outputDir,
				Language:  tmpl.Language,
				Template:  tmpl.Name,
			}

			if err := Forge(p, cfg, false); err != nil {
				t.Fatalf("Forge() error = %v", err)
			}

			if err := PostProcess(p, &PostProcessOptions{GitInit: false}, false); err != nil {
				t.Fatalf("PostProcess() error = %v", err)
			}

			cmd := exec.Command(verify[0], verify[1:]...)
			cmd.Dir = outputDir
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("%v failed: %v\n%s", verify, err, out)
			}
		})
	}
}
