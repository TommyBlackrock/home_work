package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

var ErrInvalidEnvName = errors.New("invalid environment variable name")

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read env directory %q: %w", dir, err)
	}

	env := make(Environment, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.Contains(name, "=") {
			return nil, fmt.Errorf("%w: %q contains '='", ErrInvalidEnvName, name)
		}

		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read env file %q: %w", path, err)
		}

		value, needRemove := parseEnvValue(data)
		env[name] = EnvValue{
			Value:      value,
			NeedRemove: needRemove,
		}
	}

	return env, nil
}

func parseEnvValue(data []byte) (string, bool) {
	if len(data) == 0 {
		return "", true
	}

	line, _, _ := bytes.Cut(data, []byte{'\n'})
	line = bytes.ReplaceAll(line, []byte{0x00}, []byte{'\n'})

	return strings.TrimRight(string(line), " \t"), false
}
