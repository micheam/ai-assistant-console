package main

import (
	"strings"
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
			input:    "gpt-4.1",
			expected: ModelSpec{Provider: "", ModelName: "gpt-4.1"},
		},
		{
			name:     "qualified name with provider",
			input:    "openai:gpt-4.1",
			expected: ModelSpec{Provider: "openai", ModelName: "gpt-4.1"},
		},
		{
			name:     "anthropic provider",
			input:    "anthropic:claude-haiku-5-5",
			expected: ModelSpec{Provider: "anthropic", ModelName: "claude-haiku-5-5"},
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
		{"openai", "gpt-4.1", "openai:gpt-4.1"},
		{"anthropic", "claude-haiku-5-5", "anthropic:claude-haiku-5-5"},
		{"groq", "llama-3.3-70b-versatile", "groq:llama-3.3-70b-versatile"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := QualifiedName(tt.provider, tt.modelName)
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
			spec:            "openai:gpt-4.1",
			defaultProvider: "",
			wantProvider:    "openai",
			wantModelName:   "gpt-4.1",
			wantFound:       true,
		},
		{
			name:            "explicit provider anthropic",
			spec:            "anthropic:claude-haiku-5-5",
			defaultProvider: "",
			wantProvider:    "anthropic",
			wantModelName:   "claude-haiku-5-5",
			wantFound:       true,
		},
		{
			name:            "simple name auto-detect anthropic",
			spec:            "claude-haiku-5-5",
			defaultProvider: "",
			wantProvider:    "anthropic",
			wantModelName:   "claude-haiku-5-5",
			wantFound:       true,
		},
		{
			name:            "simple name auto-detect openai",
			spec:            "gpt-4.1",
			defaultProvider: "",
			wantProvider:    "openai",
			wantModelName:   "gpt-4.1",
			wantFound:       true,
		},
		{
			name:            "invalid provider in spec",
			spec:            "invalid:gpt-4.1",
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
			spec:            "claude-haiku-5-5",
			defaultProvider: "anthropic",
			wantProvider:    "anthropic",
			wantModelName:   "claude-haiku-5-5",
			wantFound:       true,
		},
		{
			name:          "alias with explicit provider",
			spec:          "anthropic:fable",
			wantProvider:  "anthropic",
			wantModelName: "claude-fable-5-1",
			wantFound:     true,
		},
		{
			name:          "alias auto-detect anthropic",
			spec:          "fable",
			wantProvider:  "anthropic",
			wantModelName: "claude-fable-5-1",
			wantFound:     true,
		},
		{
			name:      "alias is scoped to its provider",
			spec:      "openai:fable",
			wantFound: false,
		},
		{
			name:            "alias resolves in provider other than default",
			spec:            "luna",
			defaultProvider: "anthropic",
			wantProvider:    "openai",
			wantModelName:   "gpt-6-luna",
			wantFound:       true,
		},
		{
			name:            "default provider ignored when model not supported",
			spec:            "gpt-4.1",
			defaultProvider: "anthropic",
			wantProvider:    "openai",
			wantModelName:   "gpt-4.1",
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

func TestResolveModelSpec(t *testing.T) {
	tests := []struct {
		name            string
		spec            string
		defaultProvider string
		wantProvider    string
		wantModelName   string
		wantFellBack    bool
		wantErr         bool
	}{
		{
			name:          "available model: no fallback",
			spec:          "anthropic:claude-opus-5-5",
			wantProvider:  "anthropic",
			wantModelName: "claude-opus-5-5",
		},
		{
			name:          "unavailable model with explicit provider",
			spec:          "anthropic:claude-opus-4-6",
			wantProvider:  "anthropic",
			wantModelName: "claude-haiku-5-5",
			wantFellBack:  true,
		},
		{
			name:            "explicit provider wins over default provider",
			spec:            "cerebras:llama3.1-8b",
			defaultProvider: "openai",
			wantProvider:    "cerebras",
			wantModelName:   "gpt-oss-120b",
			wantFellBack:    true,
		},
		{
			name:            "unavailable simple name uses default provider",
			spec:            "gpt-4o",
			defaultProvider: "openai",
			wantProvider:    "openai",
			wantModelName:   "gpt-6-luna",
			wantFellBack:    true,
		},
		{
			name:          "unavailable simple name without default provider",
			spec:          "claude-sonnet-4-6",
			wantProvider:  "anthropic",
			wantModelName: "claude-haiku-5-5",
			wantFellBack:  true,
		},
		{
			name:    "unknown provider in spec",
			spec:    "antropic:claude-haiku-5-5",
			wantErr: true,
		},
		{
			name:            "unknown default provider",
			spec:            "unknown-model",
			defaultProvider: "unknown",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, modelName, fellBack, err := resolveModelSpec(tt.spec, tt.defaultProvider)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantProvider, provider, "provider mismatch")
			assert.Equal(t, tt.wantModelName, modelName, "modelName mismatch")
			assert.Equal(t, tt.wantFellBack, fellBack, "fellBack mismatch")
		})
	}
}

func TestProviders_DefaultModelIsSupported(t *testing.T) {
	for _, p := range providers {
		t.Run(p.name, func(t *testing.T) {
			desc, found := p.describe(p.defaultModel)
			assert.True(t, found, "default model %q of %s is unknown", p.defaultModel, p.name)
			assert.False(t, strings.HasPrefix(desc, "[Deprecated]"),
				"default model %q of %s is deprecated", p.defaultModel, p.name)
		})
	}
}

func TestListItemView_String(t *testing.T) {
	tests := []struct {
		name string
		view listItemView
		want string
	}{
		{"plain", listItemView{QualifiedName: "groq:llama"}, "groq:llama"},
		{"aliases", listItemView{QualifiedName: "openai:gpt-6-sol", Aliases: []string{"sol"}}, "openai:gpt-6-sol (sol)"},
		{"selected with aliases", listItemView{QualifiedName: "anthropic:claude-fable-5-1", Aliases: []string{"fable"}, Selected: true}, "anthropic:claude-fable-5-1 (fable) *"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.view.String())
		})
	}
}

func TestAliasesOf(t *testing.T) {
	assert.Equal(t, []string{"fable"}, aliasesOf("anthropic", "claude-fable-5-1"))
	assert.Equal(t, []string{}, aliasesOf("anthropic", "claude-fable-5"))
	assert.Equal(t, []string{}, aliasesOf("groq", "llama-3.3-70b-versatile"))
	assert.Equal(t, []string{}, aliasesOf("unknown", "x"))
}
