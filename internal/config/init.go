package config

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Interactive walks the user through building a Config via prompts,
// reading from r and writing prompt text to w.
func Interactive(r io.Reader, w io.Writer) (Config, error) {
	reader := bufio.NewReader(r)

	baseModule, err := prompt(reader, w, "Base module path (e.g. github.com/you)", "local")
	if err != nil {
		return Config{}, err
	}

	return Config{
		BaseModule: baseModule,
	}, nil
}

// prompt asks a question, shows a default in brackets, and returns the
// user's answer — or the default if they just press enter.
func prompt(reader *bufio.Reader, w io.Writer, question, defaultVal string) (string, error) {
	fmt.Fprintf(w, "%s [%s]: ", question, defaultVal)

	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultVal, nil
	}
	return input, nil
}
