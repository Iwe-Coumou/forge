package forger

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fatih/color"
)

type PostProcessOptions struct {
	GitInit bool
}

func PostProcess(p *Project, opts *PostProcessOptions, verbose bool) error {
	if verbose {
		color.Cyan("Post processing...")
	}

	lang, err := lookupLanguage(p.Language)
	if err != nil {
		return err
	}

	if err := lang.PostProcess(p.OutputDir, verbose); err != nil {
		return err
	}

	if opts.GitInit {
		if err := runIn(p.OutputDir, verbose, "git", "init"); err != nil {
			return fmt.Errorf("git init: %w", err)
		}
		if err := runIn(p.OutputDir, verbose, "git", "add", "-A"); err != nil {
			return fmt.Errorf("git add: %w", err)
		}
		if err := runIn(p.OutputDir, verbose, "git", "commit", "-m", "Initial commit from Forge"); err != nil {
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

	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		return nil
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w\n%s", name, err, output)
	}
	return nil
}
