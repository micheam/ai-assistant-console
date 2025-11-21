package main

import (
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

func TestBuildMessageText_WithSource(t *testing.T) {
	prompt := "explain this"
	source := "func main() {}"

	result := buildMessageText(prompt, source)

	if !strings.Contains(result, prompt) {
		t.Errorf("result should contain prompt: %s", result)
	}
	if !strings.Contains(result, source) {
		t.Errorf("result should contain source: %s", result)
	}
	if !strings.Contains(result, "<source>") {
		t.Errorf("result should contain <source> tag: %s", result)
	}
}

func TestBuildMessageText_WithoutSource(t *testing.T) {
	prompt := "explain this"
	source := ""

	result := buildMessageText(prompt, source)

	if result != prompt {
		t.Errorf("expected %q, got %q", prompt, result)
	}
}
