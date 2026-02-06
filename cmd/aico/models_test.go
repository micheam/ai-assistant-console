package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseModelSpec(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ModelSpec
	}{
		{
			name:     "simple model name",
			input:    "gpt-4o",
			expected: ModelSpec{Provider: "", ModelName: "gpt-4o"},
		},
		{
			name:     "qualified name with provider",
			input:    "openai:gpt-4o",
			expected: ModelSpec{Provider: "openai", ModelName: "gpt-4o"},
		},
		{
			name:     "anthropic provider",
			input:    "anthropic:claude-haiku-4-5",
			expected: ModelSpec{Provider: "anthropic", ModelName: "claude-haiku-4-5"},
		},
		{
			name:     "groq provider",
			input:    "groq:llama-3.3-70b-versatile",
			expected: ModelSpec{Provider: "groq", ModelName: "llama-3.3-70b-versatile"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: ModelSpec{Provider: "", ModelName: ""},
		},
		{
			name:     "colon only",
			input:    ":",
			expected: ModelSpec{Provider: "", ModelName: ""},
		},
		{
			name:     "model with hyphen",
			input:    "claude-sonnet-4-5",
			expected: ModelSpec{Provider: "", ModelName: "claude-sonnet-4-5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseModelSpec(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQualifiedName(t *testing.T) {
	tests := []struct {
		provider  string
		modelName string
		expected  string
	}{
		{"openai", "gpt-4o", "openai:gpt-4o"},
		{"anthropic", "claude-haiku-4-5", "anthropic:claude-haiku-4-5"},
		{"groq", "llama-3.3-70b-versatile", "groq:llama-3.3-70b-versatile"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := QualifiedName(tt.provider, tt.modelName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestListItemView_String(t *testing.T) {
	tests := []struct {
		name     string
		view     listItemView
		expected string
	}{
		{
			name:     "normal model",
			view:     listItemView{QualifiedName: "openai:gpt-5.2"},
			expected: "openai:gpt-5.2",
		},
		{
			name:     "selected model",
			view:     listItemView{QualifiedName: "openai:gpt-5.2", Selected: true},
			expected: "openai:gpt-5.2 *",
		},
		{
			name:     "deprecated model",
			view:     listItemView{QualifiedName: "openai:gpt-4o", Deprecated: true},
			expected: "openai:gpt-4o [deprecated]",
		},
		{
			name:     "deprecated and selected model",
			view:     listItemView{QualifiedName: "openai:gpt-4o", Deprecated: true, Selected: true},
			expected: "openai:gpt-4o [deprecated] *",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.view.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectProviderByModelSpec(t *testing.T) {
	tests := []struct {
		name            string
		spec            string
		defaultProvider string
		wantProvider    string
		wantModelName   string
		wantFound       bool
	}{
		{
			name:            "explicit provider openai",
			spec:            "openai:gpt-4o",
			defaultProvider: "",
			wantProvider:    "openai",
			wantModelName:   "gpt-4o",
			wantFound:       true,
		},
		{
			name:            "explicit provider anthropic",
			spec:            "anthropic:claude-haiku-4-5",
			defaultProvider: "",
			wantProvider:    "anthropic",
			wantModelName:   "claude-haiku-4-5",
			wantFound:       true,
		},
		{
			name:            "simple name auto-detect anthropic",
			spec:            "claude-haiku-4-5",
			defaultProvider: "",
			wantProvider:    "anthropic",
			wantModelName:   "claude-haiku-4-5",
			wantFound:       true,
		},
		{
			name:            "simple name auto-detect openai",
			spec:            "gpt-4o",
			defaultProvider: "",
			wantProvider:    "openai",
			wantModelName:   "gpt-4o",
			wantFound:       true,
		},
		{
			name:            "invalid provider in spec",
			spec:            "invalid:gpt-4o",
			defaultProvider: "",
			wantProvider:    "",
			wantModelName:   "",
			wantFound:       false,
		},
		{
			name:            "unknown model",
			spec:            "unknown-model",
			defaultProvider: "",
			wantProvider:    "",
			wantModelName:   "",
			wantFound:       false,
		},
		{
			name:            "default provider used when model found",
			spec:            "claude-haiku-4-5",
			defaultProvider: "anthropic",
			wantProvider:    "anthropic",
			wantModelName:   "claude-haiku-4-5",
			wantFound:       true,
		},
		{
			name:            "default provider ignored when model not supported",
			spec:            "gpt-4o",
			defaultProvider: "anthropic",
			wantProvider:    "openai",
			wantModelName:   "gpt-4o",
			wantFound:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, modelName, found := detectProviderByModelSpec(tt.spec, tt.defaultProvider)
			assert.Equal(t, tt.wantProvider, provider, "provider mismatch")
			assert.Equal(t, tt.wantModelName, modelName, "modelName mismatch")
			assert.Equal(t, tt.wantFound, found, "found mismatch")
		})
	}
}
