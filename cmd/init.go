package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Iwe-Coumou/forge/internal/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var useDefaults bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Forge configuration",
	Long:  "Sets up Forge's config file (base module path, etc). Run interactively, or use --default to skip prompts.",
	Args:  cobra.NoArgs,
	RunE:  runInit,
}

func init() {
	initCmd.Flags().BoolVarP(&useDefaults, "default", "d", false, "skip prompts and use default values")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	var cfg config.Config
	var err error

	if useDefaults {
		cfg = config.Default()
	} else {
		cfg, err = config.Interactive(os.Stdin, os.Stdout)
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}
	}

	exists, err := config.Exists()
	if err != nil {
		return fmt.Errorf("checking existing config: %w", err)
	}

	if exists {
		overwrite, err := confirmOverwrite(os.Stdin, os.Stdout)
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}
		if !overwrite {
			color.Yellow("Aborted.")
			return nil
		}
	}

	if err := config.Save(&cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	color.Green("Forge initialized.")
	return nil
}

func confirmOverwrite(r io.Reader, w io.Writer) (bool, error) {
	reader := bufio.NewReader(r)
	fmt.Fprint(w, "Config already exists. Overwrite? [y/N]: ")

	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes", nil
}
