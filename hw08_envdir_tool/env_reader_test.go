package main

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestReadDir(t *testing.T) {
	t.Parallel()

	env, err := ReadDir("testdata/env")
	if err != nil {
		t.Fatalf("ReadDir returned error: %v", err)
	}

	expected := Environment{
		"BAR":   {Value: "bar"},
		"EMPTY": {Value: ""},
		"FOO":   {Value: "   foo\nwith new line"},
		"HELLO": {Value: `"hello"`},
		"UNSET": {NeedRemove: true},
	}
	if !reflect.DeepEqual(expected, env) {
		t.Fatalf("unexpected environment:\nwant: %#v\n got: %#v", expected, env)
	}
}

func TestReadDirInvalidEnvName(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "BAD=NAME"), "value")

	_, err := ReadDir(dir)
	if !errors.Is(err, ErrInvalidEnvName) {
		t.Fatalf("expected ErrInvalidEnvName, got %v", err)
	}
}

func TestReadDirMissingDirectory(t *testing.T) {
	t.Parallel()

	_, err := ReadDir(filepath.Join(t.TempDir(), "missing"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected os.ErrNotExist, got %v", err)
	}
}

func TestReadDirSkipsSubdirectories(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "FOO"), "bar")
	if err := os.Mkdir(filepath.Join(dir, "NESTED"), 0o755); err != nil {
		t.Fatalf("create nested directory: %v", err)
	}

	env, err := ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir returned error: %v", err)
	}

	expected := Environment{"FOO": {Value: "bar"}}
	if !reflect.DeepEqual(expected, env) {
		t.Fatalf("unexpected environment:\nwant: %#v\n got: %#v", expected, env)
	}
}

func writeTestFile(t *testing.T, path, data string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("write test file %s: %v", path, err)
	}
}
