package config

import (
	"bytes"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/require"
)

func TestLoadFromReader_Models(t *testing.T) {
	t.Run("parses [models] alongside the top-level model key", func(t *testing.T) {
		src := `
model = "claude-haiku-4-5"

[models."claude-opus-5-5"]
effort = "high"

[models."anthropic:claude-opus-5"]
effort = "anthropic:xhigh"
max_tokens = 65536

[models."gpt-5.6-sol"]
effort = "high"
`
		conf, err := loadFromReader(strings.NewReader(src))
		require.NoError(t, err)
		require.Equal(t, "claude-haiku-4-5", conf.Model)
		require.Equal(t, ModelSettings{Effort: "high"}, conf.Models["claude-opus-5-5"])
		require.Equal(t, ModelSettings{Effort: "anthropic:xhigh", MaxTokens: 65536}, conf.Models["anthropic:claude-opus-5"])
		require.Equal(t, ModelSettings{Effort: "high"}, conf.Models["gpt-5.6-sol"])
	})

	t.Run("no [models]: leaves the map nil", func(t *testing.T) {
		conf, err := loadFromReader(strings.NewReader(`model = "claude-haiku-4-5"`))
		require.NoError(t, err)
		require.Nil(t, conf.Models)
	})

	t.Run("rejects an effort outside the portable set", func(t *testing.T) {
		src := `
[models."claude-opus-5-5"]
effort = "hgih"
`
		_, err := loadFromReader(strings.NewReader(src))
		require.Error(t, err)
		require.Contains(t, err.Error(), `models."claude-opus-5-5"`)
		require.Contains(t, err.Error(), `"hgih"`)
	})

	t.Run("rejects an unprefixed provider-specific effort", func(t *testing.T) {
		src := `
[models."claude-opus-5-5"]
effort = "xhigh"
`
		_, err := loadFromReader(strings.NewReader(src))
		require.Error(t, err)
	})

	t.Run("accepts any prefixed effort at load time", func(t *testing.T) {
		src := `
[models."claude-opus-5-5"]
effort = "anthropic:whatever"
`
		_, err := loadFromReader(strings.NewReader(src))
		require.NoError(t, err)
	})

	t.Run("rejects a negative max_tokens", func(t *testing.T) {
		src := `
[models."claude-opus-5-5"]
max_tokens = -1
`
		_, err := loadFromReader(strings.NewReader(src))
		require.Error(t, err)
		require.Contains(t, err.Error(), "max_tokens")
	})
}

func TestConfig_ModelSettings(t *testing.T) {
	conf := &Config{Models: map[string]ModelSettings{
		"claude-opus-5":           {Effort: "low"},
		"anthropic:claude-opus-5": {Effort: "high"},
		"gpt-5.6-sol":             {Effort: "medium"},
	}}

	t.Run("qualified key wins over the simple key", func(t *testing.T) {
		key, s, ok := conf.ModelSettings("anthropic", "claude-opus-5")
		require.True(t, ok)
		require.Equal(t, "anthropic:claude-opus-5", key)
		require.Equal(t, "high", s.Effort)
	})

	t.Run("falls back to the simple key", func(t *testing.T) {
		key, s, ok := conf.ModelSettings("openai", "gpt-5.6-sol")
		require.True(t, ok)
		require.Equal(t, "gpt-5.6-sol", key)
		require.Equal(t, "medium", s.Effort)
	})

	t.Run("missing model", func(t *testing.T) {
		_, _, ok := conf.ModelSettings("anthropic", "claude-haiku-4-5")
		require.False(t, ok)
	})

	t.Run("nil map", func(t *testing.T) {
		_, _, ok := (&Config{}).ModelSettings("anthropic", "claude-opus-5")
		require.False(t, ok)
	})
}

func TestModelSettings_ResolveEffort(t *testing.T) {
	tests := []struct {
		name     string
		effort   string
		provider string
		want     string
		wantErr  bool
	}{
		{name: "empty", effort: "", provider: "anthropic", want: ""},
		{name: "portable value passes through", effort: "high", provider: "openai", want: "high"},
		{name: "matching prefix is stripped", effort: "anthropic:xhigh", provider: "anthropic", want: "xhigh"},
		{name: "mismatched prefix is an error", effort: "openai:minimal", provider: "anthropic", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ModelSettings{Effort: tt.effort}.ResolveEffort(tt.provider)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestDefaultConfig_EncodeOmitsModels(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, toml.NewEncoder(&buf).Encode(DefaultConfig()))
	require.NotContains(t, buf.String(), "[models")
}
