package forger

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// TestTemplateMetadataIsConsistent checks every template's yaml against the
// folder it lives in, and that its language is actually registered — the two
// ways a template can be added correctly but wired up wrong.
func TestTemplateMetadataIsConsistent(t *testing.T) {
	templates, err := ListTemplates()
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
			lang, tmpl, err := ParseTemplateID(tt.arg)

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
	detail, err := InspectTemplate("go", "cli_cobra")
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
	detail, err := InspectTemplate("python", "cli")
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
	if _, err := InspectTemplate("go", "does_not_exist"); err == nil {
		t.Fatal("InspectTemplate() = nil error, want error for an unknown template")
	}
}

func TestTemplateInfoID(t *testing.T) {
	got := TemplateInfo{Language: "go", Name: "cli_cobra"}.ID()
	if got != "go/cli_cobra" {
		t.Errorf("ID() = %q, want %q", got, "go/cli_cobra")
	}
}
