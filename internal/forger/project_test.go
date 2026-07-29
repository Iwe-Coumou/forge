package forger

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProjectValidate(t *testing.T) {
	tmp := t.TempDir()

	nonEmptyDir := filepath.Join(tmp, "nonempty")
	if err := os.MkdirAll(nonEmptyDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nonEmptyDir, "file.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	emptyDir := filepath.Join(tmp, "empty")
	if err := os.MkdirAll(emptyDir, 0o755); err != nil {
		t.Fatal(err)
	}

	missingDir := filepath.Join(tmp, "missing")

	tests := []struct {
		name    string
		project Project
		wantErr bool
	}{
		{
			name:    "missing name",
			project: Project{ModulePath: "m", OutputDir: missingDir, Template: "cli_cobra"},
			wantErr: true,
		},
		{
			name:    "missing module path",
			project: Project{Name: "n", OutputDir: missingDir, Template: "cli_cobra"},
			wantErr: true,
		},
		{
			name:    "missing output dir",
			project: Project{Name: "n", ModulePath: "m", Template: "cli_cobra"},
			wantErr: true,
		},
		{
			name:    "missing template",
			project: Project{Name: "n", ModulePath: "m", OutputDir: missingDir},
			wantErr: true,
		},
		{
			name:    "unknown template",
			project: Project{Name: "n", ModulePath: "m", OutputDir: missingDir, Template: "does_not_exist"},
			wantErr: true,
		},
		{
			name:    "output dir does not exist",
			project: Project{Name: "n", ModulePath: "m", OutputDir: missingDir, Template: "cli_cobra"},
			wantErr: false,
		},
		{
			name:    "output dir exists and is empty",
			project: Project{Name: "n", ModulePath: "m", OutputDir: emptyDir, Template: "cli_cobra"},
			wantErr: false,
		},
		{
			name:    "output dir exists and is not empty",
			project: Project{Name: "n", ModulePath: "m", OutputDir: nonEmptyDir, Template: "cli_cobra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.project.validate()
			if tt.wantErr && err == nil {
				t.Fatalf("validate() = nil, want error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("validate() = %v, want nil", err)
			}
		})
	}
}
