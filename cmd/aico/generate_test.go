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
