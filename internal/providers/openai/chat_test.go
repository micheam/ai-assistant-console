package openai

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"micheam.com/aico/internal/assistant"
)

func TestBuildChatRequest_GenerationOptions(t *testing.T) {
	ctx := context.Background()
	msgs := []assistant.Message{assistant.NewUserMessage(assistant.NewTextContent("hi"))}

	t.Run("zero options: omits reasoning_effort and max_completion_tokens", func(t *testing.T) {
		req, err := BuildChatRequest(ctx, "gpt-5.6-sol", nil, assistant.GenerationOptions{}, msgs)
		require.NoError(t, err)
		data, err := json.Marshal(req)
		require.NoError(t, err)
		require.NotContains(t, string(data), `"reasoning_effort"`)
		require.NotContains(t, string(data), `"max_completion_tokens"`)
		require.NotContains(t, string(data), `"max_tokens"`)
	})

	t.Run("set options: sends reasoning_effort and max_completion_tokens", func(t *testing.T) {
		opts := assistant.GenerationOptions{Effort: "minimal", MaxTokens: 4096}
		req, err := BuildChatRequest(ctx, "gpt-5.6-sol", nil, opts, msgs)
		require.NoError(t, err)
		data, err := json.Marshal(req)
		require.NoError(t, err)
		require.Contains(t, string(data), `"reasoning_effort":"minimal"`)
		require.Contains(t, string(data), `"max_completion_tokens":4096`)
		require.NotContains(t, string(data), `"max_tokens"`)
	})
}

func TestAliases_PointToSupportedModels(t *testing.T) {
	for alias, target := range Aliases() {
		t.Run(alias, func(t *testing.T) {
			m, ok := selectModel(target)
			require.True(t, ok, "alias %q points to unknown model %q", alias, target)
			require.False(t, strings.HasPrefix(m.Description(), "[Deprecated]"),
				"alias %q points to deprecated model %q", alias, target)
		})
	}
}
