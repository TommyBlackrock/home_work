package main

import (
	"errors"
	"os"
	"os/exec"
	"sort"
	"strings"
)

const (
	exitCodeOK              = 0
	exitCodeUsageError      = 1
	exitCodeCannotExecute   = 126
	exitCodeCommandNotFound = 127
)

func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 || cmd[0] == "" {
		return exitCodeUsageError
	}
	//nolint:gosec // envdir intentionally executes the command provided by the user
	command := exec.Command(cmd[0], cmd[1:]...)
	command.Env = patchEnvironment(os.Environ(), env)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	if err := command.Run(); err != nil {
		return commandErrorExitCode(err)
	}

	return exitCodeOK
}

func patchEnvironment(base []string, env Environment) []string {
	if len(env) == 0 {
		return append([]string{}, base...)
	}

	result := make([]string, 0, len(base)+len(env))
	seen := make(map[string]struct{}, len(env))

	for _, item := range base {
		name, _, ok := strings.Cut(item, "=")
		if !ok {
			continue
		}

		value, exists := env[name]
		if !exists {
			result = append(result, item)
			continue
		}

		seen[name] = struct{}{}
		if !value.NeedRemove {
			result = append(result, name+"="+value.Value)
		}
	}

	names := make([]string, 0, len(env))
	for name, value := range env {
		if _, exists := seen[name]; exists || value.NeedRemove {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		result = append(result, name+"="+env[name].Value)
	}

	return result
}

func commandErrorExitCode(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}

	if errors.Is(err, exec.ErrNotFound) || errors.Is(err, os.ErrNotExist) {
		return exitCodeCommandNotFound
	}
	if errors.Is(err, os.ErrPermission) {
		return exitCodeCannotExecute
	}

	return exitCodeUsageError
}
