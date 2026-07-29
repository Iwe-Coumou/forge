package config

import (
	"bytes"
	"strings"
	"testing"
)

// answers builds stdin for Interactive from one line per prompt, in order:
// base module, author, email, license, git init.
func answers(lines ...string) *strings.Reader {
	return strings.NewReader(strings.Join(lines, "\n") + "\n")
}

func TestInteractive_UsesProvidedValues(t *testing.T) {
	r := answers("github.com/someone", "Ada Lovelace", "ada@example.com", "MIT", "y")
	var w bytes.Buffer

	cfg, err := Interactive(r, &w)
	if err != nil {
		t.Fatalf("Interactive() error = %v", err)
	}

	if got := cfg.LanguageSetting("go", "base_module"); got != "github.com/someone" {
		t.Errorf("go base_module = %q, want %q", got, "github.com/someone")
	}
	if cfg.Author != "Ada Lovelace" {
		t.Errorf("Author = %q, want %q", cfg.Author, "Ada Lovelace")
	}
	if cfg.Email != "ada@example.com" {
		t.Errorf("Email = %q, want %q", cfg.Email, "ada@example.com")
	}
	if cfg.License != "MIT" {
		t.Errorf("License = %q, want %q", cfg.License, "MIT")
	}
	if !cfg.GitInit {
		t.Error("GitInit = false, want true")
	}

	for _, want := range []string{"Base module path", "Author name", "Author email", "License", "Initialize git"} {
		if !strings.Contains(w.String(), want) {
			t.Errorf("prompt output = %q, want it to ask for %q", w.String(), want)
		}
	}
}

func TestInteractive_BlankInputUsesDefaults(t *testing.T) {
	r := answers("", "", "", "", "")
	var w bytes.Buffer

	cfg, err := Interactive(r, &w)
	if err != nil {
		t.Fatalf("Interactive() error = %v", err)
	}

	if got := cfg.LanguageSetting("go", "base_module"); got != "local" {
		t.Errorf("go base_module = %q, want default %q", got, "local")
	}
	if cfg.Author != "" || cfg.Email != "" || cfg.License != "" {
		t.Errorf("Author/Email/License = %q/%q/%q, want all empty", cfg.Author, cfg.Email, cfg.License)
	}
	if cfg.GitInit {
		t.Error("GitInit = true, want false by default")
	}
}

func TestInteractive_TrimsWhitespace(t *testing.T) {
	r := answers("  github.com/spacey  ", "  Ada  ", "", "", "")
	var w bytes.Buffer

	cfg, err := Interactive(r, &w)
	if err != nil {
		t.Fatalf("Interactive() error = %v", err)
	}

	if got := cfg.LanguageSetting("go", "base_module"); got != "github.com/spacey" {
		t.Errorf("go base_module = %q, want trimmed %q", got, "github.com/spacey")
	}
	if cfg.Author != "Ada" {
		t.Errorf("Author = %q, want trimmed %q", cfg.Author, "Ada")
	}
}

// TestInteractive_ShortInput covers piping fewer lines than there are
// prompts: the remaining prompts must fall back to their defaults rather
// than failing on EOF.
func TestInteractive_ShortInput(t *testing.T) {
	r := strings.NewReader("github.com/someone\n")
	var w bytes.Buffer

	cfg, err := Interactive(r, &w)
	if err != nil {
		t.Fatalf("Interactive() error = %v", err)
	}

	if got := cfg.LanguageSetting("go", "base_module"); got != "github.com/someone" {
		t.Errorf("go base_module = %q, want %q", got, "github.com/someone")
	}
	if cfg.Author != "" {
		t.Errorf("Author = %q, want empty when the prompt got no input", cfg.Author)
	}
}

func TestInteractive_GitInitAnswers(t *testing.T) {
	tests := []struct {
		answer string
		want   bool
	}{
		{answer: "y", want: true},
		{answer: "yes", want: true},
		{answer: "Y", want: true},
		{answer: "n", want: false},
		{answer: "no", want: false},
		{answer: "", want: false},
		{answer: "nonsense", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.answer, func(t *testing.T) {
			var w bytes.Buffer
			cfg, err := Interactive(answers("", "", "", "", tt.answer), &w)
			if err != nil {
				t.Fatalf("Interactive() error = %v", err)
			}
			if cfg.GitInit != tt.want {
				t.Errorf("GitInit = %v for answer %q, want %v", cfg.GitInit, tt.answer, tt.want)
			}
		})
	}
}

func TestInteractive_DoesNotWriteDeprecatedKey(t *testing.T) {
	cfg, err := Interactive(answers("github.com/someone", "", "", "", ""), &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Interactive() error = %v", err)
	}
	if cfg.BaseModule != "" {
		t.Errorf("BaseModule = %q, want empty — new configs must not write the deprecated key", cfg.BaseModule)
	}
}
