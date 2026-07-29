package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Iwe-Coumou/forge/v2/internal/config"
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
	// One reader shared by both prompts — a second bufio.Reader over the same
	// stdin would buffer ahead and swallow input the other one needs.
	reader := bufio.NewReader(os.Stdin)

	// Confirm before asking anything, so an aborted run costs no answers.
	exists, err := config.Exists()
	if err != nil {
		return fmt.Errorf("checking existing config: %w", err)
	}
	if exists {
		overwrite, err := confirmOverwrite(reader, os.Stdout)
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}
		if !overwrite {
			color.Yellow("Aborted.")
			return nil
		}
	}

	var cfg config.Config
	if useDefaults {
		cfg = config.Default()
	} else {
		cfg, err = config.Interactive(reader, os.Stdout)
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}
	}

	if err := config.Save(&cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	color.Green("Forge initialized.")
	return nil
}

// confirmOverwrite asks before replacing an existing config. Anything other
// than an explicit yes — including EOF — leaves the config alone.
func confirmOverwrite(reader *bufio.Reader, w io.Writer) (bool, error) {
	fmt.Fprint(w, "Config already exists. Overwrite? [y/N]: ")

	input, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}

	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes", nil
}
