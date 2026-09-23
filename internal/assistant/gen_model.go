package assistant

import (
	"context"
	"errors"
	"iter"
)

// Sentinel errors a GenerativeModel's GenerateContentStream may report
// instead of a normal terminal Usage item, when the turn ended in a state
// unsafe to treat as a complete response (e.g. a tool_use proposal that was
// cut off mid-input). Callers must not persist the assistant message for a
// turn that failed with one of these.
var (
	// ErrMaxTokens indicates the response was cut off by the max_tokens
	// limit before the model finished its turn.
	ErrMaxTokens = errors.New("generation stopped at max_tokens")

	// ErrRefusal indicates the model declined to continue the turn
	// (stop_reason "refusal"). Content produced before the refusal, if any,
	// must not be treated as a complete response.
	ErrRefusal = errors.New("generation stopped: refusal")

	// ErrTruncatedToolUse indicates a tool_use content block's input was
	// not valid JSON when its block completed, which can happen when the
	// turn is cut off mid-parameter.
	ErrTruncatedToolUse = errors.New("tool_use input is not complete JSON")
)

type ModelDescriptor interface {
	Name() string
	Description() string
	Provider() string
}

// GenerativeModel represents a generative model.
type GenerativeModel interface {
	ModelDescriptor
	SetSystemInstruction(...*TextContent)
	GenerateContent(context.Context, ...Message) (*GenerateContentResponse, error)
	GenerateContentStream(context.Context, ...Message) (iter.Seq2[*GenerateContentResponse, error], error)
}

type GenerateContentResponse struct {
	Content MessageContent

	// Usage carries token accounting for the generation. It is only populated
	// on the terminal chunk of a stream (where Content is nil), or always for
	// a non-streaming response.
	Usage *Usage
}

// Usage reports token accounting for a single generation, normalized across
// providers that use different mechanisms for prompt caching (e.g. OpenAI's
// fully automatic caching vs. Anthropic's explicit cache_control breakpoints).
type Usage struct {
	InputTokens       int `json:"input_tokens"`
	OutputTokens      int `json:"output_tokens"`
	CachedInputTokens int `json:"cached_input_tokens"` // subset of InputTokens served from a prompt cache
	CacheWriteTokens  int `json:"cache_write_tokens"`  // tokens newly written to a prompt cache this call (0 if unsupported/not reported)
}

// CacheHitRate returns the percentage of InputTokens served from a prompt cache.
func (u *Usage) CacheHitRate() float64 {
	if u == nil || u.InputTokens == 0 {
		return 0
	}
	return 100 * float64(u.CachedInputTokens) / float64(u.InputTokens)
}
