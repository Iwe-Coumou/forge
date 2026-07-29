package config

import (
	"bytes"
	"strings"
	"testing"
)

func TestInteractive_UsesProvidedValue(t *testing.T) {
	r := strings.NewReader("github.com/someone\n")
	var w bytes.Buffer

	cfg, err := Interactive(r, &w)
	if err != nil {
		t.Fatalf("Interactive() error = %v", err)
	}
	if cfg.BaseModule != "github.com/someone" {
		t.Errorf("BaseModule = %q, want %q", cfg.BaseModule, "github.com/someone")
	}
	if !strings.Contains(w.String(), "Base module path") {
		t.Errorf("prompt output = %q, want it to ask for the base module path", w.String())
	}
}

func TestInteractive_BlankInputUsesDefault(t *testing.T) {
	r := strings.NewReader("\n")
	var w bytes.Buffer

	cfg, err := Interactive(r, &w)
	if err != nil {
		t.Fatalf("Interactive() error = %v", err)
	}
	if cfg.BaseModule != "local" {
		t.Errorf("BaseModule = %q, want default %q", cfg.BaseModule, "local")
	}
}

func TestInteractive_TrimsWhitespace(t *testing.T) {
	r := strings.NewReader("  github.com/spacey  \n")
	var w bytes.Buffer

	cfg, err := Interactive(r, &w)
	if err != nil {
		t.Fatalf("Interactive() error = %v", err)
	}
	if cfg.BaseModule != "github.com/spacey" {
		t.Errorf("BaseModule = %q, want trimmed %q", cfg.BaseModule, "github.com/spacey")
	}
}
