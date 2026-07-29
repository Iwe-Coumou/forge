package config

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Interactive walks the user through building a Config via prompts,
// reading from r and writing prompt text to w. Every prompt accepts a blank
// line to take its default, so pressing enter through it is valid.
func Interactive(r io.Reader, w io.Writer) (Config, error) {
	// Reuse an existing bufio.Reader rather than wrapping it: a second layer
	// of buffering over the same stream can swallow the caller's input.
	reader, ok := r.(*bufio.Reader)
	if !ok {
		reader = bufio.NewReader(r)
	}

	baseModule, err := prompt(reader, w, "Base module path (e.g. github.com/you)", "local")
	if err != nil {
		return Config{}, err
	}

	author, err := prompt(reader, w, "Author name", "")
	if err != nil {
		return Config{}, err
	}

	email, err := prompt(reader, w, "Author email", "")
	if err != nil {
		return Config{}, err
	}

	license, err := prompt(reader, w, "License (e.g. MIT)", "")
	if err != nil {
		return Config{}, err
	}

	gitInit, err := promptBool(reader, w, "Initialize git in new projects?", false)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Author:  author,
		Email:   email,
		License: license,
		GitInit: gitInit,
		Languages: map[string]map[string]string{
			"go": {"base_module": baseModule},
		},
	}, nil
}

// prompt asks a question, shows a default in brackets when there is one, and
// returns the user's answer — or the default if they just press enter.
func prompt(reader *bufio.Reader, w io.Writer, question, defaultVal string) (string, error) {
	if defaultVal == "" {
		fmt.Fprintf(w, "%s: ", question)
	} else {
		fmt.Fprintf(w, "%s [%s]: ", question, defaultVal)
	}

	input, err := readLine(reader)
	if err != nil {
		return "", err
	}

	if input == "" {
		return defaultVal, nil
	}
	return input, nil
}

// promptBool asks a yes/no question. Anything unrecognised takes the default.
func promptBool(reader *bufio.Reader, w io.Writer, question string, defaultVal bool) (bool, error) {
	hint := "y/N"
	if defaultVal {
		hint = "Y/n"
	}
	fmt.Fprintf(w, "%s [%s]: ", question, hint)

	input, err := readLine(reader)
	if err != nil {
		return false, err
	}

	switch strings.ToLower(input) {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return defaultVal, nil
	}
}

// readLine reads one trimmed line. Input that ends without a trailing newline
// is still returned, so piping fewer lines than there are prompts leaves the
// remaining ones on their defaults instead of failing.
func readLine(reader *bufio.Reader) (string, error) {
	input, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(input), nil
}
