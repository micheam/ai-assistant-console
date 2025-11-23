package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadSource_WithContent(t *testing.T) {
	input := "hello world"
	r := strings.NewReader(input)

	result, err := readSource(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != input {
		t.Errorf("expected %q, got %q", input, result)
	}
}

func TestReadSource_EmptyInput(t *testing.T) {
	r := strings.NewReader("")

	result, err := readSource(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestResolveSource_DirectString(t *testing.T) {
	input := "function foo() { return 42; }"

	result, err := resolveSource(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "<source>\n" + input + "\n</source>"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestResolveSource_FromFile(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	content := "package main\n\nfunc main() {}\n"
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	result, err := resolveSource("@" + tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "<source file=") {
		t.Errorf("expected result to contain '<source file=', got %q", result)
	}
	if !strings.Contains(result, content) {
		t.Errorf("expected result to contain file content, got %q", result)
	}
	if !strings.Contains(result, "</source>") {
		t.Errorf("expected result to contain '</source>', got %q", result)
	}
}

func TestResolveSource_FileNotFound(t *testing.T) {
	_, err := resolveSource("@/nonexistent/file.go")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
