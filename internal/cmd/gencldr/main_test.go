package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/romshark/icumsg/internal/test"
)

func TestGenerate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cldr_gen.go")
	test.RequireNoErr(t, generate(path, "cldr"))

	actual, err := os.ReadFile(path)
	test.RequireNoErr(t, err)
	expect, err := os.ReadFile("../../cldr/cldr_gen.go")
	test.RequireNoErr(t, err)
	if !bytes.Equal(expect, actual) {
		t.Fatal("internal/cldr/cldr_gen.go is outdated, run go generate ./internal/cldr")
	}
}

func TestGenerateErrKeepsFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cldr_gen.go")
	const existing = "package cldr\n"
	test.RequireNoErr(t, os.WriteFile(path, []byte(existing), 0o644))

	// The invalid package name makes formatting fail.
	if err := generate(path, "invalid-name"); err == nil {
		t.Fatal("expected error")
	}

	actual, err := os.ReadFile(path)
	test.RequireNoErr(t, err)
	test.RequireEqual(t, existing, string(actual))

	entries, err := os.ReadDir(dir)
	test.RequireNoErr(t, err)
	test.RequireEqual(t, 1, len(entries), "temporary file left behind")
}

func TestGenerateErrRename(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cldr_gen.go")
	// Renaming the temporary file fails because a directory is in the way.
	test.RequireNoErr(t, os.Mkdir(path, 0o755))
	test.RequireNoErr(t, os.WriteFile(filepath.Join(path, "file"), nil, 0o644))

	if err := generate(path, "cldr"); err == nil {
		t.Fatal("expected error")
	}

	entries, err := os.ReadDir(path)
	test.RequireNoErr(t, err)
	test.RequireEqual(t, 1, len(entries), "directory in the way changed")

	entries, err = os.ReadDir(dir)
	test.RequireNoErr(t, err)
	test.RequireEqual(t, 1, len(entries), "temporary file left behind")
}

func TestGenerateErrWrite(t *testing.T) {
	dir := t.TempDir()
	// Writing the temporary file fails because its directory is missing.
	path := filepath.Join(dir, "missing", "cldr_gen.go")

	if err := generate(path, "cldr"); err == nil {
		t.Fatal("expected error")
	}

	entries, err := os.ReadDir(dir)
	test.RequireNoErr(t, err)
	test.RequireEqual(t, 0, len(entries), "file created")
}
