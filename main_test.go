package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadMappingsHappyPath(t *testing.T) {
	dir := t.TempDir()
	snippets := filepath.Join(dir, "snippets")
	if err := os.Mkdir(snippets, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(snippets, "world.txt"), []byte("there"), 0o644); err != nil {
		t.Fatal(err)
	}

	mappings, err := loadMappings(snippets)
	if err != nil {
		t.Fatal(err)
	}

	if mappings["world"] != "there" {
		t.Errorf("got %q, wanted %q", mappings["world"], "there")
	}
}

func TestLoadMappingsEmptyDir(t *testing.T) {
	dir := t.TempDir()

	mappings, err := loadMappings(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(mappings) != 0 {
		t.Errorf("expected empty mappings, got %v", mappings)
	}
}

func TestLoadMappingsDuplicateStem(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "foo.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "foo.md"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := loadMappings(dir)
	if err == nil {
		t.Fatal("expected error for duplicate stem")
	}
	if !strings.Contains(err.Error(), `duplicate stem "foo"`) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSortedStemsLongestFirst(t *testing.T) {
	mappings := map[string]string{
		"FOO":     "1",
		"FOO_BAR": "2",
		"X":       "3",
	}

	stems := sortedStems(mappings)
	if stems[0] != "FOO_BAR" {
		t.Errorf("expected FOO_BAR first, got %q", stems[0])
	}
}

func TestHappyPath(t *testing.T) {
	_ = os.Mkdir("out", 0o755)

	snippetsDir := "out/snippets"
	targetDir := "out/target"
	if err := os.MkdirAll(snippetsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(snippetsDir, "world.txt"), []byte("there"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "hello.txt"), []byte("Hello world\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("INPUT_SOURCE_DIR", snippetsDir)
	t.Setenv("INPUT_INCLUDE", "out/target/**")
	t.Setenv("INPUT_EXCLUDE", "")
	t.Setenv("GITHUB_OUTPUT", "out/output.txt")

	main()

	data, err := os.ReadFile(filepath.Join(targetDir, "hello.txt"))
	if err != nil {
		t.Fatal(err)
	}

	want := "Hello there\n"
	if string(data) != want {
		t.Errorf("got %q, wanted %q", data, want)
	}

	output, err := os.ReadFile("out/output.txt")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(output), "modifiedFiles=1") {
		t.Errorf("expected modifiedFiles=1 in output, got %q", output)
	}
}
