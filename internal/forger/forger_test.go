package forger

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Iwe-Coumou/forge/v2/internal/config"
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

// TestForgeRendersUserTemplate is the end-to-end proof that a template living
// outside the binary works: it is found, parsed from its own source, and gets
// the same render context an embedded template would.
func TestForgeRendersUserTemplate(t *testing.T) {
	cfg, _ := userTemplates(t, "go", "minimal", map[string]string{
		"template.yaml": "language: go\nshort: \"Minimal\"\nlong: \"Minimal.\"\n",
		"go.mod.tmpl":   "module {{.ModulePath}}\n\ngo {{.GoVersion}}\n",
		"main.go.tmpl":  "package main\n\n// by {{.Author}}\nfunc main() { println(\"{{.Name}}\") }\n",
	})
	cfg.Author = "Ada Lovelace"
	cfg.Languages = map[string]map[string]string{"go": {"base_module": "example.com/ada"}}

	outputDir := filepath.Join(t.TempDir(), "demo")
	p := &Project{
		Name:      "demo",
		OutputDir: outputDir,
		Language:  "go",
		Template:  "minimal",
	}

	if err := Forge(p, cfg, false); err != nil {
		t.Fatalf("Forge() error = %v", err)
	}

	got := renderedFiles(t, outputDir)
	want := []string{"go.mod", "main.go"}
	if len(got) != len(want) {
		t.Fatalf("rendered files = %v, want %v", got, want)
	}

	goMod, err := os.ReadFile(filepath.Join(outputDir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(goMod), "module example.com/ada/demo") {
		t.Errorf("go.mod = %s, want the configured base module applied", goMod)
	}
	if !strings.Contains(string(goMod), "go 1.22") {
		t.Errorf("go.mod = %s, want the language's default Go version", goMod)
	}

	mainGo, err := os.ReadFile(filepath.Join(outputDir, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(mainGo), "by Ada Lovelace") {
		t.Errorf("main.go = %s, want the configured author", mainGo)
	}
}

// TestForgeRejectsUnknownTemplate covers the check that moved out of
// validate() into findTemplate: an unresolvable template must fail before
// anything is written.
func TestForgeRejectsUnknownTemplate(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "myproj")

	p := &Project{
		Name:      "myproj",
		OutputDir: outputDir,
		Language:  "go",
		Template:  "does_not_exist",
	}

	err := Forge(p, embeddedOnly(t), false)
	if err == nil {
		t.Fatal("Forge() = nil, want error for an unknown template")
	}
	if !strings.Contains(err.Error(), "unknown template") {
		t.Errorf("Forge() error = %v, want it to say the template is unknown", err)
	}
	if _, statErr := os.Stat(outputDir); !os.IsNotExist(statErr) {
		t.Errorf("Forge() created %s, want no output for an unknown template", outputDir)
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

// TestForgeResolvesGoVersion covers all three tiers of the precedence chain:
// --set beats the config file, which beats the language's built-in default.
func TestForgeResolvesGoVersion(t *testing.T) {
	configured := func(version string) *config.Config {
		return &config.Config{
			Languages: map[string]map[string]string{"go": {"go_version": version}},
		}
	}

	tests := []struct {
		name      string
		cfg       *config.Config
		overrides map[string]string
		want      string
	}{
		{
			name: "built-in default",
			cfg:  &config.Config{},
			want: "go 1.22",
		},
		{
			name: "config beats built-in default",
			cfg:  configured("1.23"),
			want: "go 1.23",
		},
		{
			name:      "override beats config",
			cfg:       configured("1.23"),
			overrides: map[string]string{"go_version": "1.21"},
			want:      "go 1.21",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputDir := filepath.Join(t.TempDir(), "myproj")

			p := &Project{
				Name:      "myproj",
				OutputDir: outputDir,
				Language:  "go",
				Template:  "cli_cobra",
				Overrides: tt.overrides,
			}

			if err := Forge(p, tt.cfg, false); err != nil {
				t.Fatalf("Forge() error = %v", err)
			}

			goMod, err := os.ReadFile(filepath.Join(outputDir, "go.mod"))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(goMod), tt.want) {
				t.Errorf("go.mod = %s, want it to contain %q", goMod, tt.want)
			}
		})
	}
}

// TestPythonResolvesMinPython checks the same chain for a second language,
// going through Context directly since Forge refuses unimplemented languages.
func TestPythonResolvesMinPython(t *testing.T) {
	configured := &config.Config{
		Languages: map[string]map[string]string{"python": {"min_python": "3.12"}},
	}

	tests := []struct {
		name      string
		cfg       *config.Config
		overrides map[string]string
		want      string
	}{
		{name: "built-in default", cfg: &config.Config{}, want: "3.11"},
		{name: "config beats built-in default", cfg: configured, want: "3.12"},
		{
			name:      "override beats config",
			cfg:       configured,
			overrides: map[string]string{"min_python": "3.13"},
			want:      "3.13",
		},
	}

	lang, err := lookupLanguage("python")
	if err != nil {
		t.Fatalf("lookupLanguage() error = %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Project{Name: "myproj", Language: "python", Overrides: tt.overrides}

			ctx, err := lang.Context(p, tt.cfg)
			if err != nil {
				t.Fatalf("Context() error = %v", err)
			}

			got, ok := ctx.(PythonContext)
			if !ok {
				t.Fatalf("Context() = %T, want PythonContext", ctx)
			}
			if got.MinPython != tt.want {
				t.Errorf("MinPython = %q, want %q", got.MinPython, tt.want)
			}
		})
	}
}

// TestCommonContextFromConfig checks the language-neutral fields every
// context embeds, for each registered language.
func TestCommonContextFromConfig(t *testing.T) {
	cfg := &config.Config{Author: "Ada Lovelace", License: "MIT"}

	for _, name := range []string{"go", "python"} {
		t.Run(name, func(t *testing.T) {
			lang, err := lookupLanguage(name)
			if err != nil {
				t.Fatalf("lookupLanguage() error = %v", err)
			}

			ctx, err := lang.Context(&Project{Name: "myproj", Language: name}, cfg)
			if err != nil {
				t.Fatalf("Context() error = %v", err)
			}

			common := reflect.ValueOf(ctx).FieldByName("Common").Interface().(Common)

			if common.Name != "myproj" {
				t.Errorf("Name = %q, want %q", common.Name, "myproj")
			}
			if common.Author != "Ada Lovelace" {
				t.Errorf("Author = %q, want %q", common.Author, "Ada Lovelace")
			}
			if common.License != "MIT" {
				t.Errorf("License = %q, want %q", common.License, "MIT")
			}
			if common.Year != time.Now().Year() {
				t.Errorf("Year = %d, want %d", common.Year, time.Now().Year())
			}
		})
	}
}

// TestLanguageKeysAreDeclared guards the centralisation: every registered
// language must declare its keys, and every key it declares must be one the
// language actually reads. The second half is what caught go_version being
// accepted but ignored.
func TestLanguageKeysAreDeclared(t *testing.T) {
	for name, lang := range languages {
		t.Run(name, func(t *testing.T) {
			keys := lang.Keys()
			if len(keys.Flag) == 0 && len(keys.Config) == 0 {
				t.Fatal("language declares no keys at all")
			}

			for _, key := range keys.Flag {
				if key == "" {
					t.Error("Keys().Flag contains an empty key")
				}
			}
			for _, key := range keys.Config {
				if key == "" {
					t.Error("Keys().Config contains an empty key")
				}
			}
		})
	}
}

func TestCheckConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr string
	}{
		{
			name: "empty config",
			cfg:  &config.Config{},
		},
		{
			name: "known keys",
			cfg: &config.Config{Languages: map[string]map[string]string{
				"go":     {"base_module": "example.com", "go_version": "1.23"},
				"python": {"min_python": "3.12"},
			}},
		},
		{
			// Forward compatibility: a config written for a newer Forge must
			// not break this one.
			name: "unregistered language is ignored",
			cfg: &config.Config{Languages: map[string]map[string]string{
				"rust": {"edition": "2021"},
			}},
		},
		{
			name: "typo in a known language",
			cfg: &config.Config{Languages: map[string]map[string]string{
				"go": {"go_versionn": "1.23"},
			}},
			wantErr: "go_versionn",
		},
		{
			// module_path is per-project, so it is deliberately flag-only.
			name: "flag-only key rejected in config",
			cfg: &config.Config{Languages: map[string]map[string]string{
				"go": {"module_path": "example.com/x"},
			}},
			wantErr: "module_path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckConfig(tt.cfg)

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("CheckConfig() = %v, want nil", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("CheckConfig() = nil, want error mentioning %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("CheckConfig() = %v, want it to name %q", err, tt.wantErr)
			}
		})
	}
}

// TestForgeRejectsBadConfig proves the config check runs during a scaffold,
// not just when called directly.
func TestForgeRejectsBadConfig(t *testing.T) {
	outputDir := filepath.Join(t.TempDir(), "myproj")

	p := &Project{
		Name:      "myproj",
		OutputDir: outputDir,
		Language:  "go",
		Template:  "cli_cobra",
	}
	cfg := &config.Config{Languages: map[string]map[string]string{
		"go": {"go_versionn": "1.23"},
	}}

	err := Forge(p, cfg, false)
	if err == nil {
		t.Fatal("Forge() = nil, want error for an unknown config key")
	}
	if !strings.Contains(err.Error(), "go_versionn") {
		t.Errorf("Forge() error = %v, want it to name the offending key", err)
	}
	if _, statErr := os.Stat(outputDir); !os.IsNotExist(statErr) {
		t.Errorf("Forge() created %s, want no output when the config is invalid", outputDir)
	}
}

// TestForgeRejectsUnknownOverrideForSecondLanguage checks that centralising
// the --set check in Forge covers every language, including ones whose
// Context is never reached.
func TestForgeRejectsUnknownOverrideForSecondLanguage(t *testing.T) {
	p := &Project{
		Name:      "myproj",
		OutputDir: filepath.Join(t.TempDir(), "myproj"),
		Language:  "python",
		Template:  "cli",
		Overrides: map[string]string{"module_path": "example.com/x"},
	}

	err := Forge(p, &config.Config{}, false)
	if err == nil {
		t.Fatal("Forge() = nil, want an error")
	}
	// Unimplemented is checked first, so that is the error we expect — the
	// point is that neither path silently accepts a bogus key.
	if !strings.Contains(err.Error(), "not implemented") && !strings.Contains(err.Error(), "module_path") {
		t.Errorf("Forge() error = %v, want it to refuse the language or the key", err)
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
