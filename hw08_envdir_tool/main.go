package main

import (
	"errors"
	"fmt"
	"os"
)

const minArgsCount = 2

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) < minArgsCount {
		fmt.Fprintln(os.Stderr, "Usage: go-envdir <envdir> <command> [args...]")
		return exitCodeUsageError
	}

	envDir := args[0]
	command := args[1:]

	env, err := ReadDir(envDir)
	if err != nil {
		printReadDirError(envDir, err)
		return exitCodeUsageError
	}

	return RunCmd(command, env)
}

func printReadDirError(envDir string, err error) {
	switch {
	case errors.Is(err, ErrInvalidEnvName):
		fmt.Fprintf(os.Stderr, "Error: invalid environment variable file name: %v\n", err)
	case errors.Is(err, os.ErrNotExist):
		fmt.Fprintf(os.Stderr, "Error: env directory %s does not exist\n", envDir)
	case errors.Is(err, os.ErrPermission):
		fmt.Fprintf(os.Stderr, "Error: permission denied while reading env directory %s\n", envDir)
	default:
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}
