package openai

import (
	"context"
	"encoding/json"
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
