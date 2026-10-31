package complete

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompletion_JSON(t *testing.T) {
	// These types are sent to editor clients as is, so pin their field
	// names here.
	b, err := json.Marshal(Completion{
		Candidates: []Candidate{{Text: "t"}, {Text: "u", Truncated: true}},
		Timings:    Timings{PromptTokens: 1, PromptMS: 2, PredictedTokens: 3, PredictedMS: 4, CachedTokens: 5},
	})
	require.NoError(t, err)
	assert.JSONEq(t, `{
		"candidates": [{"text": "t", "truncated": false}, {"text": "u", "truncated": true}],
		"timings": {"prompt_tokens": 1, "prompt_ms": 2, "predicted_tokens": 3, "predicted_ms": 4, "cached_tokens": 5}
	}`, string(b))
}

func TestRequest_JSON(t *testing.T) {
	var req Request
	require.NoError(t, json.Unmarshal([]byte(`{"prefix":"p","suffix":"s","indent":2,"extra":[{"file":"f","text":"t"}],"count":3}`), &req))
	assert.Equal(t, Request{Prefix: "p", Suffix: "s", Indent: 2, Extra: []Chunk{{File: "f", Text: "t"}}, Count: 3}, req)
}
