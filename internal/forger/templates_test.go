package forger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Iwe-Coumou/forge/v2/internal/config"
	"go.yaml.in/yaml/v3"
)

// embeddedOnly returns a config whose user templates directory does not
// exist, so tests see only the templates compiled into the binary.
//
// An empty Config is not good enough: TemplatesDir() falls back to
// <user config dir>/forge/templates, which may actually exist on a
// developer's machine and would leak their templates into these assertions.
func embeddedOnly(t *testing.T) *config.Config {
	t.Helper()
	return &config.Config{TemplatesDirectory: filepath.Join(t.TempDir(), "absent")}
}

// userTemplates writes a template into a temp directory and returns a config
// pointing at it, plus the directory itself so tests can assert on Source.
// Keys of files are slash-separated paths relative to the template root.
func userTemplates(t *testing.T, lang, name string, files map[string]string) (*config.Config, string) {
	t.Helper()

	dir := t.TempDir()
	root := filepath.Join(dir, lang, name)

	for rel, content := range files {
		full := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	return &config.Config{TemplatesDirectory: dir}, dir
}

// minimalTemplate is the smallest valid Go template: metadata plus one file.
func minimalTemplate() map[string]string {
	return map[string]string{
		"template.yaml": "language: go\nshort: \"Bare-bones Go program\"\nlong: \"A single main.go.\"\n",
		"main.go.tmpl":  "package main\n\n// {{.ModulePath}}\nfunc main() {}\n",
	}
}

func TestListTemplates_IncludesUserTemplates(t *testing.T) {
	cfg, dir := userTemplates(t, "go", "minimal", minimalTemplate())

	templates, err := ListTemplates(cfg)
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}

	var found *TemplateInfo
	for i := range templates {
		if templates[i].ID() == "go/minimal" {
			found = &templates[i]
		}
	}
	if found == nil {
		t.Fatalf("go/minimal not listed, got %v", templates)
	}

	if found.Source != dir {
		t.Errorf("Source = %q, want the user templates dir %q", found.Source, dir)
	}
	if found.Short != "Bare-bones Go program" {
		t.Errorf("Short = %q, want it read from the user template.yaml", found.Short)
	}

	// The embedded templates must still be there alongside it.
	for _, want := range []string{"go/cli_cobra", "python/cli"} {
		if !containsID(templates, want) {
			t.Errorf("%s missing, want embedded templates listed too (got %v)", want, templates)
		}
	}
}

func TestListTemplates_EmbeddedSourceLabel(t *testing.T) {
	templates, err := ListTemplates(embeddedOnly(t))
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}

	for _, tmpl := range templates {
		if tmpl.Source != "embedded" {
			t.Errorf("%s Source = %q, want %q", tmpl.ID(), tmpl.Source, "embedded")
		}
	}
}

// TestListTemplates_RejectsCollision pins the conservative choice: a user
// template that shadows an embedded one is an error, not a silent override.
func TestListTemplates_RejectsCollision(t *testing.T) {
	cfg, dir := userTemplates(t, "go", "cli_cobra", minimalTemplate())

	_, err := ListTemplates(cfg)
	if err == nil {
		t.Fatal("ListTemplates() = nil, want error for a template defined twice")
	}
	if !strings.Contains(err.Error(), "go/cli_cobra") {
		t.Errorf("error = %v, want it to name the colliding template", err)
	}
	if !strings.Contains(err.Error(), "embedded") || !strings.Contains(err.Error(), dir) {
		t.Errorf("error = %v, want it to name both sources", err)
	}
}

// TestListTemplates_MissingUserDirIsFine covers the common case: the default
// templates directory does not exist for most users.
func TestListTemplates_MissingUserDirIsFine(t *testing.T) {
	templates, err := ListTemplates(embeddedOnly(t))
	if err != nil {
		t.Fatalf("ListTemplates() error = %v, want an absent dir to be ignored", err)
	}
	if len(templates) == 0 {
		t.Fatal("no templates listed")
	}
}

func TestInspectTemplate_UserTemplate(t *testing.T) {
	cfg, dir := userTemplates(t, "go", "minimal", minimalTemplate())

	detail, err := InspectTemplate(cfg, "go", "minimal")
	if err != nil {
		t.Fatalf("InspectTemplate() error = %v", err)
	}

	if detail.Source != dir {
		t.Errorf("Source = %q, want %q", detail.Source, dir)
	}
	if want := []string{"main.go"}; len(detail.Files) != 1 || detail.Files[0] != want[0] {
		t.Errorf("Files = %v, want %v", detail.Files, want)
	}
	// The language comes from the registry, not from the source, so a user
	// template still gets its verify command and keys.
	if len(detail.VerifyCmd) == 0 {
		t.Error("VerifyCmd is empty, want the Go language's verify command")
	}
	if len(detail.Keys.Flag) == 0 {
		t.Error("Keys.Flag is empty, want the Go language's --set keys")
	}
}

// TestParseTemplateID_ResolvesUserTemplate checks that the CLI argument path
// resolves against user templates too, not just embedded ones.
func TestParseTemplateID_ResolvesUserTemplate(t *testing.T) {
	cfg, _ := userTemplates(t, "go", "minimal", minimalTemplate())

	lang, name, err := ParseTemplateID(cfg, "go/minimal")
	if err != nil {
		t.Fatalf("ParseTemplateID() error = %v", err)
	}
	if lang != "go" || name != "minimal" {
		t.Errorf("ParseTemplateID() = (%q, %q), want (%q, %q)", lang, name, "go", "minimal")
	}
}

// TestWalkTemplate_RejectsUnsafeDestination exercises the filepath.IsLocal
// guard. A file named exactly ".tmpl" strips to an empty destination, which is
// the one trigger reachable from a real directory source on every OS —
// fs.WalkDir cleans away "..", and Windows reserved names like NUL can't be
// created on disk to test with.
func TestWalkTemplate_RejectsUnsafeDestination(t *testing.T) {
	cfg, _ := userTemplates(t, "go", "unsafe", map[string]string{
		"template.yaml": "language: go\nshort: \"Unsafe\"\nlong: \"Unsafe.\"\n",
		".tmpl":         "nameless\n",
	})

	_, err := InspectTemplate(cfg, "go", "unsafe")
	if err == nil {
		t.Fatal("InspectTemplate() = nil, want error for a template file with no destination name")
	}
	if !strings.Contains(err.Error(), "escape the project directory") {
		t.Errorf("error = %v, want the path guard's message", err)
	}
}

func containsID(templates []TemplateInfo, id string) bool {
	for _, t := range templates {
		if t.ID() == id {
			return true
		}
	}
	return false
}

// TestTemplateMetadataIsConsistent checks every template's yaml against the
// folder it lives in, and that its language is actually registered — the two
// ways a template can be added correctly but wired up wrong.
func TestTemplateMetadataIsConsistent(t *testing.T) {
	templates, err := ListTemplates(embeddedOnly(t))
	if err != nil {
		t.Fatalf("ListTemplates() error = %v", err)
	}
	if len(templates) == 0 {
		t.Fatal("no templates found")
	}

	for _, tmpl := range templates {
		t.Run(tmpl.ID(), func(t *testing.T) {
			if _, err := lookupLanguage(tmpl.Language); err != nil {
				t.Errorf("template folder %q has no registered language: %v", tmpl.ID(), err)
			}

			data, err := templateFS.ReadFile("templates/" + tmpl.Language + "/" + tmpl.Name + "/template.yaml")
			if err != nil {
				t.Fatalf("reading template.yaml: %v", err)
			}

			var meta templateMeta
			if err := yaml.Unmarshal(data, &meta); err != nil {
				t.Fatalf("parsing template.yaml: %v", err)
			}

			if meta.Language != tmpl.Language {
				t.Errorf("template.yaml language = %q, want %q to match its folder", meta.Language, tmpl.Language)
			}
			if meta.Short == "" {
				t.Error("template.yaml has no short description")
			}
			if meta.Long == "" {
				t.Error("template.yaml has no long description")
			}
		})
	}
}

func TestParseTemplateID(t *testing.T) {
	tests := []struct {
		name     string
		arg      string
		wantLang string
		wantTmpl string
		wantErr  bool
	}{
		{name: "qualified", arg: "go/cli_cobra", wantLang: "go", wantTmpl: "cli_cobra"},
		{name: "other language", arg: "python/cli", wantLang: "python", wantTmpl: "cli"},
		{name: "backslash is normalised", arg: `go\cli_cobra`, wantLang: "go", wantTmpl: "cli_cobra"},
		{name: "bare name", arg: "cli_cobra", wantErr: true},
		{name: "unknown template", arg: "go/does_not_exist", wantErr: true},
		{name: "unknown language", arg: "rust/cli", wantErr: true},
		{name: "wrong language for template", arg: "python/cli_cobra", wantErr: true},
		{name: "empty language", arg: "/cli_cobra", wantErr: true},
		{name: "empty template", arg: "go/", wantErr: true},
		{name: "too many segments", arg: "go/cli/cobra", wantErr: true},
		{name: "empty string", arg: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lang, tmpl, err := ParseTemplateID(embeddedOnly(t), tt.arg)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseTemplateID(%q) = (%q, %q, nil), want error", tt.arg, lang, tmpl)
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseTemplateID(%q) error = %v", tt.arg, err)
			}
			if lang != tt.wantLang || tmpl != tt.wantTmpl {
				t.Errorf("ParseTemplateID(%q) = (%q, %q), want (%q, %q)",
					tt.arg, lang, tmpl, tt.wantLang, tt.wantTmpl)
			}
		})
	}
}

func TestInspectTemplate(t *testing.T) {
	detail, err := InspectTemplate(embeddedOnly(t), "go", "cli_cobra")
	if err != nil {
		t.Fatalf("InspectTemplate() error = %v", err)
	}

	if detail.ID() != "go/cli_cobra" {
		t.Errorf("ID() = %q, want %q", detail.ID(), "go/cli_cobra")
	}
	if detail.Short == "" {
		t.Error("Short is empty, want the template.yaml description")
	}
	if detail.NotImplemented != "" {
		t.Errorf("NotImplemented = %q, want empty for an implemented language", detail.NotImplemented)
	}

	// Same list TestForgeRendersTemplateFiles asserts against the rendered
	// output — if these ever disagree, inspect is lying about what forge writes.
	want := []string{"cmd/example.go", "cmd/root.go", "go.mod", "main.go"}
	if len(detail.Files) != len(want) {
		t.Fatalf("Files = %v, want %v", detail.Files, want)
	}
	for i, w := range want {
		if detail.Files[i] != w {
			t.Errorf("Files[%d] = %q, want %q (full list %v)", i, detail.Files[i], w, detail.Files)
		}
	}

	if len(detail.VerifyCmd) == 0 {
		t.Error("VerifyCmd is empty, want the language's verify command")
	}
}

func TestInspectTemplate_ReportsUnimplemented(t *testing.T) {
	detail, err := InspectTemplate(embeddedOnly(t), "python", "cli")
	if err != nil {
		t.Fatalf("InspectTemplate() error = %v", err)
	}
	if detail.NotImplemented == "" {
		t.Error("NotImplemented is empty, want a reason for a wip language")
	}
	if len(detail.Files) == 0 {
		t.Error("Files is empty, want inspect to work for unimplemented languages")
	}
}

func TestInspectTemplate_Unknown(t *testing.T) {
	if _, err := InspectTemplate(embeddedOnly(t), "go", "does_not_exist"); err == nil {
		t.Fatal("InspectTemplate() = nil error, want error for an unknown template")
	}
}

func TestTemplateInfoID(t *testing.T) {
	got := TemplateInfo{Language: "go", Name: "cli_cobra"}.ID()
	if got != "go/cli_cobra" {
		t.Errorf("ID() = %q, want %q", got, "go/cli_cobra")
	}
}
