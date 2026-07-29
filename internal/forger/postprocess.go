package forger

import (
	"fmt"
	"os/exec"

	"github.com/fatih/color"
)

type PostProcessOptions struct {
	GitInit bool
}

func PostProcess(outputDir string, opts *PostProcessOptions, verbose bool) error {
	if verbose {
		color.Cyan("Post processing...")
	}
	if err := runIn(outputDir, verbose, "go", "mod", "tidy"); err != nil {
		return fmt.Errorf("go mod tidy: %w", err)
	}

	if err := runIn(outputDir, verbose, "gofmt", "-w", "."); err != nil {
		return fmt.Errorf("gofmt: %w", err)
	}

	if opts.GitInit {
		if err := runIn(outputDir, verbose, "git", "init"); err != nil {
			return fmt.Errorf("git init: %w", err)
		}
		if err := runIn(outputDir, verbose, "git", "add", "-A"); err != nil {
			return fmt.Errorf("git add: %w", err)
		}
		if err := runIn(outputDir, verbose, "git", "commit", "-m", "Initial commit from Forge"); err != nil {
			return fmt.Errorf("git commit: %w", err)
		}
	}

	return nil
}

func runIn(dir string, verbose bool, name string, args ...string) error {
	if verbose {
		color.Magenta("running: %s %v (in %s)\n", name, args, dir)
	}

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w\n%s", name, err, output)
	}
	return nil
}
