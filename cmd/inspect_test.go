package cmd

import (
	"bytes"
	"os"
	"path"
	"sort"
	"strings"
	"testing"
)

// leafPaths flattens a tree back into the slash-separated paths of its leaves,
// so buildTree can be checked by round-tripping its input.
func leafPaths(node *treeNode, prefix string) []string {
	if len(node.children) == 0 {
		return []string{prefix}
	}

	var out []string
	for name, child := range node.children {
		out = append(out, leafPaths(child, path.Join(prefix, name))...)
	}
	return out
}

func TestBuildTree_RoundTripsPaths(t *testing.T) {
	tests := []struct {
		name  string
		paths []string
	}{
		{name: "single file", paths: []string{"main.go"}},
		{name: "flat", paths: []string{"go.mod", "main.go"}},
		{
			name:  "shared directory",
			paths: []string{"cmd/example.go", "cmd/root.go", "go.mod", "main.go"},
		},
		{
			name:  "deeply nested",
			paths: []string{"internal/a/b/c/deep.go", "internal/a/other.go"},
		},
		{
			name:  "sibling directories",
			paths: []string{"api/handler.go", "web/index.html"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := leafPaths(buildTree(tt.paths), "")
			sort.Strings(got)

			want := append([]string(nil), tt.paths...)
			sort.Strings(want)

			if strings.Join(got, "|") != strings.Join(want, "|") {
				t.Errorf("leaves = %v, want %v", got, want)
			}
		})
	}
}

// TestBuildTree_CollapsesSharedPrefix is the behaviour the tree exists for:
// two files in the same directory must produce one directory node, not two.
func TestBuildTree_CollapsesSharedPrefix(t *testing.T) {
	root := buildTree([]string{"cmd/example.go", "cmd/root.go", "main.go"})

	if len(root.children) != 2 {
		t.Fatalf("top level has %d children, want 2 (cmd, main.go)", len(root.children))
	}

	cmdNode, ok := root.children["cmd"]
	if !ok {
		t.Fatal("no cmd node at the top level")
	}
	if len(cmdNode.children) != 2 {
		t.Errorf("cmd has %d children, want 2", len(cmdNode.children))
	}

	mainNode, ok := root.children["main.go"]
	if !ok {
		t.Fatal("no main.go node at the top level")
	}
	if len(mainNode.children) != 0 {
		t.Errorf("main.go has %d children, want 0 — files are leaves", len(mainNode.children))
	}
}

func TestBuildTree_Empty(t *testing.T) {
	root := buildTree(nil)
	if len(root.children) != 0 {
		t.Errorf("children = %d, want 0 for no paths", len(root.children))
	}
}

// captureStdout swaps os.Stdout for a pipe while fn runs. printTree writes via
// fmt.Printf, so this is enough to assert on its output.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	fn()

	os.Stdout = old
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

// TestPrintTree_Connectors pins the drawing rules: a non-final entry gets ├──
// and its children keep the │ gutter; the final entry gets └── and its
// children get blank space.
func TestPrintTree_Connectors(t *testing.T) {
	paths := []string{"cmd/example.go", "cmd/root.go", "go.mod", "main.go"}

	got := captureStdout(t, func() {
		printTree(buildTree(paths), "")
	})

	want := strings.Join([]string{
		"├── cmd",
		"│   ├── example.go",
		"│   └── root.go",
		"├── go.mod",
		"└── main.go",
		"",
	}, "\n")

	if got != want {
		t.Errorf("printTree() output:\n%s\nwant:\n%s", got, want)
	}
}

func TestPrintTree_HonoursPrefix(t *testing.T) {
	got := captureStdout(t, func() {
		printTree(buildTree([]string{"main.go"}), "  ")
	})

	if want := "  └── main.go\n"; got != want {
		t.Errorf("printTree() = %q, want %q", got, want)
	}
}

func TestPrintTree_LastChildDropsGutter(t *testing.T) {
	// web is last at the top level, so its child must be indented with spaces
	// rather than a continuing │.
	got := captureStdout(t, func() {
		printTree(buildTree([]string{"api/a.go", "web/b.go"}), "")
	})

	want := strings.Join([]string{
		"├── api",
		"│   └── a.go",
		"└── web",
		"    └── b.go",
		"",
	}, "\n")

	if got != want {
		t.Errorf("printTree() output:\n%s\nwant:\n%s", got, want)
	}
}
