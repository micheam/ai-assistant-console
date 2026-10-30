// Package complete defines engine-neutral types for code
// completion: what an editor sends and what a completion backend returns.
//
// The JSON field names are the ones exchanged with editor clients, so they
// deliberately do not follow any particular backend's API.
package complete

import "context"

// Chunk is a piece of reference text from another file, given to the
// backend as additional context.
type Chunk struct {
	File string `json:"file"`
	Text string `json:"text"`
}

// Request asks for the text to insert at the cursor.
type Request struct {
	// Prefix is the text before the cursor, including the current line up
	// to the cursor.
	Prefix string `json:"prefix"`

	// Suffix is the text after the cursor, including the rest of the
	// current line.
	Suffix string `json:"suffix"`

	// Indent is the indentation of the current line, in characters.
	Indent int `json:"indent,omitempty"`

	// Extra is optional reference text from other files.
	Extra []Chunk `json:"extra,omitempty"`

	// Count is how many candidates to generate. Zero means the profile's
	// configured default.
	Count int `json:"count,omitempty"`
}

// Timings reports how the backend spent its time on one completion.
type Timings struct {
	// PromptTokens is the number of prompt tokens newly processed, and
	// PromptMS the time spent on them.
	PromptTokens int     `json:"prompt_tokens"`
	PromptMS     float64 `json:"prompt_ms"`

	// PredictedTokens is the number of generated tokens, and PredictedMS
	// the time spent generating them.
	PredictedTokens int     `json:"predicted_tokens"`
	PredictedMS     float64 `json:"predicted_ms"`

	// CachedTokens is the number of prompt tokens reused from the
	// backend's prompt cache.
	CachedTokens int `json:"cached_tokens"`
}

// Candidate is one suggested insertion.
type Candidate struct {
	// Text is the text to insert at the cursor. It may be empty.
	Text string `json:"text"`

	// Truncated reports that the backend dropped part of the context to
	// fit its context window.
	Truncated bool `json:"truncated"`
}

// Completion is the backend's answer to a Request.
type Completion struct {
	// Candidates holds the suggestions in the backend's order. It always
	// has at least one element.
	Candidates []Candidate `json:"candidates"`

	// Timings covers the whole request, however many candidates it made.
	Timings Timings `json:"timings"`
}

// Completer produces a completion for a Request.
//
// Implementations must stop work and return an error wrapping ctx.Err()
// when ctx is cancelled, since editor clients cancel a request as soon as
// the user types again.
type Completer interface {
	Complete(ctx context.Context, req Request) (*Completion, error)
}
