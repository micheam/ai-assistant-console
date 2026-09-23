package assistant

// ToolDefinition describes a client-side tool that a GenerativeModel may
// call. It is provider-agnostic: providers that support tool use (currently
// only the Anthropic provider, via ToolCapable) translate it into their own
// request shape.
type ToolDefinition struct {
	// Name is how the tool is invoked in a tool_use block and referenced in
	// a following tool_result.
	Name string

	// Description tells the model what the tool does and, importantly,
	// when to call it. It is part of the model's request context, so
	// changing it changes the byte-identical prefix that preserved-thinking
	// accounts require across a session's later turns; treat edits to a
	// registered tool's Description or InputSchema as a breaking change for
	// sessions that already used it.
	Description string

	// InputSchema is a JSON Schema object (as a map, matching the
	// provider's raw-JSON tool definition shape) describing the tool's
	// input.
	InputSchema map[string]any
}

// ToolCapable is implemented by GenerativeModel providers that accept
// client-side tool definitions. Not every provider supports tool use, so
// this is a separate, optional interface rather than part of
// GenerativeModel itself.
type ToolCapable interface {
	SetTools(tools ...ToolDefinition)
}
