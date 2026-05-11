package main

import (
	"reflect"
	"testing"
)

func TestRunCmd(t *testing.T) {
	t.Parallel()

	code := RunCmd([]string{"sh", "-c", "exit 7"}, nil)
	if code != 7 {
		t.Fatalf("expected exit code 7, got %d", code)
	}
}

func TestRunCmdWithEnvironment(t *testing.T) {
	t.Setenv("REMOVE_ME", "old")

	env := Environment{
		"FOO":       {Value: "bar"},
		"REMOVE_ME": {NeedRemove: true},
	}

	code := RunCmd([]string{"sh", "-c", `[ "$FOO" = "bar" ] && [ -z "${REMOVE_ME+x}" ]`}, env)
	if code != exitCodeOK {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}

func TestRunCmdEmptyCommand(t *testing.T) {
	t.Parallel()

	code := RunCmd(nil, nil)
	if code != exitCodeUsageError {
		t.Fatalf("expected usage exit code, got %d", code)
	}
}

func TestPatchEnvironment(t *testing.T) {
	t.Parallel()

	base := []string{
		"FOO=old",
		"KEEP=1",
		"REMOVE_ME=old",
		"EMPTY=old",
	}
	env := Environment{
		"ADDED":     {Value: "yes"},
		"EMPTY":     {Value: ""},
		"FOO":       {Value: "new"},
		"REMOVE_ME": {NeedRemove: true},
	}

	got := patchEnvironment(base, env)
	expected := []string{
		"FOO=new",
		"KEEP=1",
		"EMPTY=",
		"ADDED=yes",
	}
	if !reflect.DeepEqual(expected, got) {
		t.Fatalf("unexpected environment:\nwant: %#v\n got: %#v", expected, got)
	}
}
