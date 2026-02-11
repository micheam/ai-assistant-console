package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"micheam.com/aico/internal/assistant"
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

func TestUserMessage_WithSource(t *testing.T) {
	source := "<source>\nhello world\n</source>"
	prompt := "summarize this"

	userContents := make([]assistant.MessageContent, 0, 2)
	userContents = append(userContents, assistant.NewTextContent(source))
	userContents = append(userContents, assistant.NewTextContent(prompt))
	userMsg := assistant.NewUserMessage(userContents...)

	contents := userMsg.GetContents()
	if len(contents) != 2 {
		t.Fatalf("expected 2 contents, got %d", len(contents))
	}
	src, ok := contents[0].(*assistant.TextContent)
	if !ok {
		t.Fatalf("expected first content to be *TextContent")
	}
	if src.Text != source {
		t.Errorf("expected source text %q, got %q", source, src.Text)
	}
	msg, ok := contents[1].(*assistant.TextContent)
	if !ok {
		t.Fatalf("expected second content to be *TextContent")
	}
	if msg.Text != prompt {
		t.Errorf("expected prompt text %q, got %q", prompt, msg.Text)
	}
}

func TestUserMessage_WithoutSource(t *testing.T) {
	prompt := "hello"

	userContents := make([]assistant.MessageContent, 0, 1)
	userContents = append(userContents, assistant.NewTextContent(prompt))
	userMsg := assistant.NewUserMessage(userContents...)

	contents := userMsg.GetContents()
	if len(contents) != 1 {
		t.Fatalf("expected 1 content, got %d", len(contents))
	}
	msg, ok := contents[0].(*assistant.TextContent)
	if !ok {
		t.Fatalf("expected content to be *TextContent")
	}
	if msg.Text != prompt {
		t.Errorf("expected prompt text %q, got %q", prompt, msg.Text)
	}
}
