package assistant

import (
	"encoding/json"
	"net/url"
)

type Message interface {
	GetAuthor() MessageAuthor
	GetContents() []MessageContent
}

// UserMessage represents a message from the user.
//
// Example:
//
//	{
//	  "author": "user",
//	  "contents": [
//	    {"text": "Hello, how are you?"},
//	    {"url": "https://example.com/image.jpg"}
//	  ]
//	}
type UserMessage struct {
	Contents []MessageContent `json:"contents"`
}

var (
	_ Message          = (*UserMessage)(nil)
	_ json.Marshaler   = (*UserMessage)(nil)
	_ json.Unmarshaler = (*UserMessage)(nil)
)

// NewUserMessage creates a new user message.
func NewUserMessage(contents ...MessageContent) *UserMessage {
	return &UserMessage{Contents: contents}
}

func (u UserMessage) GetAuthor() MessageAuthor {
	return MessageAuthorUser
}

func (u UserMessage) GetContents() []MessageContent {
	if u.Contents == nil {
		return []MessageContent{}
	}
	return u.Contents
}

func (u UserMessage) MarshalJSON() ([]byte, error) {
	type alias UserMessage
	return json.Marshal(&struct {
		Author MessageAuthor `json:"author"`
		*alias
	}{
		Author: MessageAuthorUser,
		alias:  (*alias)(&u),
	})
}

func (u *UserMessage) UnmarshalJSON(data []byte) error {
	var aux struct {
		Author   MessageAuthor     `json:"author"`
		Contents []json.RawMessage `json:"contents"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Unmarshal each content by detecting its type
	u.Contents = make([]MessageContent, 0, len(aux.Contents))
	for _, raw := range aux.Contents {
		content, err := decodeContent(raw)
		if err != nil {
			return err
		}
		if content != nil {
			u.Contents = append(u.Contents, content)
		}
	}

	return nil
}

// AssistantMessage represents a message from the assistant.
//
// Example:
//
//	{
//	  "author": "assistant",
//	  "contents": [
//	    {"text": "I'm fine, thank you!"}
//	  ]
//	}
type AssistantMessage struct {
	Contents []MessageContent `json:"contents"`
}

var (
	_ Message          = (*AssistantMessage)(nil)
	_ json.Marshaler   = (*AssistantMessage)(nil)
	_ json.Unmarshaler = (*AssistantMessage)(nil)
)

// NewAssistantMessage creates a new assistant message.
func NewAssistantMessage(contents ...MessageContent) *AssistantMessage {
	return &AssistantMessage{Contents: contents}
}

func (a AssistantMessage) GetAuthor() MessageAuthor {
	return MessageAuthorAssistant
}

func (a AssistantMessage) GetContents() []MessageContent {
	if a.Contents == nil {
		return []MessageContent{}
	}
	return a.Contents
}

func (a AssistantMessage) MarshalJSON() ([]byte, error) {
	type alias AssistantMessage
	return json.Marshal(&struct {
		Author MessageAuthor `json:"author"`
		*alias
	}{
		Author: MessageAuthorAssistant,
		alias:  (*alias)(&a),
	})
}

func (a *AssistantMessage) UnmarshalJSON(data []byte) error {
	var aux struct {
		Author   MessageAuthor     `json:"author"`
		Contents []json.RawMessage `json:"contents"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Unmarshal each content by detecting its type
	a.Contents = make([]MessageContent, 0, len(aux.Contents))
	for _, raw := range aux.Contents {
		content, err := decodeContent(raw)
		if err != nil {
			return err
		}
		if content != nil {
			a.Contents = append(a.Contents, content)
		}
	}

	return nil
}

// decodeContent unmarshals a single content element by detecting its type
// from the wrapper key it carries (tool_use, tool_result, thinking,
// redacted_thinking, text, url, in that order). An element matching none of
// these keys is unknown and is skipped: decodeContent returns (nil, nil).
func decodeContent(raw json.RawMessage) (MessageContent, error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}

	switch {
	case has(m, "tool_use"):
		var c ToolUseContent
		if err := json.Unmarshal(raw, &c); err != nil {
			return nil, err
		}
		return &c, nil
	case has(m, "tool_result"):
		var c ToolResultContent
		if err := json.Unmarshal(raw, &c); err != nil {
			return nil, err
		}
		return &c, nil
	case has(m, "thinking"):
		var c ThinkingContent
		if err := json.Unmarshal(raw, &c); err != nil {
			return nil, err
		}
		return &c, nil
	case has(m, "redacted_thinking"):
		var c RedactedThinkingContent
		if err := json.Unmarshal(raw, &c); err != nil {
			return nil, err
		}
		return &c, nil
	case has(m, "text"):
		var c TextContent
		if err := json.Unmarshal(raw, &c); err != nil {
			return nil, err
		}
		return &c, nil
	case has(m, "url"):
		var c URLImageContent
		if err := json.Unmarshal(raw, &c); err != nil {
			return nil, err
		}
		return &c, nil
	default:
		return nil, nil
	}
}

func has(m map[string]json.RawMessage, key string) bool {
	_, ok := m[key]
	return ok
}

// TODO: Remove this
type MessageAuthor string

const (
	MessageAuthorAssistant MessageAuthor = "assistant"
	MessageAuthorUser      MessageAuthor = "user"
)

var (
	_ json.Marshaler   = (*MessageAuthor)(nil)
	_ json.Unmarshaler = (*MessageAuthor)(nil)
)

func (ma *MessageAuthor) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*ma = MessageAuthor(s)
	return nil
}

func (ma MessageAuthor) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(ma))
}

type MessageContent interface {
	isMessageContent()
}

// TextContent represents a text message.
// TextContent holds the pharse of the message.
//
// Example:
//
//	{ text: "Hello, how can I help you?" }
type TextContent struct {
	Text string `json:"text"`
}

var (
	_ MessageContent   = (*TextContent)(nil)
	_ json.Marshaler   = (*TextContent)(nil)
	_ json.Unmarshaler = (*TextContent)(nil)
)

func (t *TextContent) isMessageContent() {}

func NewTextContent(text string) *TextContent {
	return &TextContent{Text: text}
}

func (t *TextContent) MarshalJSON() ([]byte, error) {
	type alias TextContent
	return json.Marshal(&struct {
		*alias
	}{
		alias: (*alias)(t),
	})
}

func (t *TextContent) UnmarshalJSON(data []byte) error {
	type alias TextContent
	var aux alias
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	*t = TextContent(aux)
	return nil
}

// URLImageContent represents an image message.
//
// URLImageContent holds the URL of the image.
//
// Example:
//
//	{ url: "https://example.com/image.jpg" }
type URLImageContent struct {
	URL url.URL `json:"url"`
}

var (
	_ MessageContent   = (*URLImageContent)(nil)
	_ json.Marshaler   = (*URLImageContent)(nil)
	_ json.Unmarshaler = (*URLImageContent)(nil)
)

func (u *URLImageContent) isMessageContent() {}

func NewURLImageContent(url url.URL) *URLImageContent {
	return &URLImageContent{URL: url}
}

func (u *URLImageContent) MarshalJSON() ([]byte, error) {
	m := make(map[string]string)
	m["url"] = u.URL.String()
	return json.Marshal(m)
}

func (u *URLImageContent) UnmarshalJSON(data []byte) error {
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	if urlStr, ok := m["url"]; ok {
		parsedURL, err := url.Parse(urlStr)
		if err != nil {
			return err
		}
		u.URL = *parsedURL
	}
	return nil
}

// ToolUseContent represents a request from the model to call a client-side
// tool. Applying the call (or not) is entirely up to the client; it is never
// executed automatically by aico.
//
// Example:
//
//	{ "tool_use": {"id": "toolu_01..", "name": "propose_edit", "input": {...}} }
type ToolUseContent struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

var (
	_ MessageContent   = (*ToolUseContent)(nil)
	_ json.Marshaler   = (*ToolUseContent)(nil)
	_ json.Unmarshaler = (*ToolUseContent)(nil)
)

func (t *ToolUseContent) isMessageContent() {}

type toolUseContentJSON struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

func (t ToolUseContent) MarshalJSON() ([]byte, error) {
	input := t.Input
	if len(input) == 0 {
		input = json.RawMessage("{}")
	}
	return json.Marshal(map[string]toolUseContentJSON{
		"tool_use": {ID: t.ID, Name: t.Name, Input: input},
	})
}

func (t *ToolUseContent) UnmarshalJSON(data []byte) error {
	var wrapper struct {
		ToolUse toolUseContentJSON `json:"tool_use"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return err
	}
	t.ID = wrapper.ToolUse.ID
	t.Name = wrapper.ToolUse.Name
	t.Input = wrapper.ToolUse.Input
	return nil
}

// ToolResultContent represents the outcome of a tool_use call, sent back to
// the model in a subsequent user message.
//
// Example:
//
//	{ "tool_result": {"tool_use_id": "toolu_01..", "content": "...", "is_error": false} }
type ToolResultContent struct {
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error"`
}

var (
	_ MessageContent   = (*ToolResultContent)(nil)
	_ json.Marshaler   = (*ToolResultContent)(nil)
	_ json.Unmarshaler = (*ToolResultContent)(nil)
)

func (t *ToolResultContent) isMessageContent() {}

type toolResultContentJSON struct {
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error"`
}

func (t ToolResultContent) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]toolResultContentJSON{
		"tool_result": {ToolUseID: t.ToolUseID, Content: t.Content, IsError: t.IsError},
	})
}

func (t *ToolResultContent) UnmarshalJSON(data []byte) error {
	var wrapper struct {
		ToolResult toolResultContentJSON `json:"tool_result"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return err
	}
	t.ToolUseID = wrapper.ToolResult.ToolUseID
	t.Content = wrapper.ToolResult.Content
	t.IsError = wrapper.ToolResult.IsError
	return nil
}

// ThinkingContent represents a model-internal reasoning block. Its text is
// normally empty (see the Anthropic API's thinking display setting) but the
// signature must still be persisted and replayed unchanged, or a later
// request that continues a tool_use turn will be rejected.
//
// Example:
//
//	{ "thinking": {"thinking": "", "signature": "Er..."} }
type ThinkingContent struct {
	Thinking  string `json:"thinking"`
	Signature string `json:"signature"`
}

var (
	_ MessageContent   = (*ThinkingContent)(nil)
	_ json.Marshaler   = (*ThinkingContent)(nil)
	_ json.Unmarshaler = (*ThinkingContent)(nil)
)

func (t *ThinkingContent) isMessageContent() {}

type thinkingContentJSON struct {
	Thinking  string `json:"thinking"`
	Signature string `json:"signature"`
}

func (t ThinkingContent) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]thinkingContentJSON{
		"thinking": {Thinking: t.Thinking, Signature: t.Signature},
	})
}

func (t *ThinkingContent) UnmarshalJSON(data []byte) error {
	var wrapper struct {
		Thinking thinkingContentJSON `json:"thinking"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return err
	}
	t.Thinking = wrapper.Thinking.Thinking
	t.Signature = wrapper.Thinking.Signature
	return nil
}

// RedactedThinkingContent represents a thinking block whose content was
// withheld by the API. Like ThinkingContent, it must be persisted and
// replayed unchanged.
//
// Example:
//
//	{ "redacted_thinking": {"data": "..."} }
type RedactedThinkingContent struct {
	Data string `json:"data"`
}

var (
	_ MessageContent   = (*RedactedThinkingContent)(nil)
	_ json.Marshaler   = (*RedactedThinkingContent)(nil)
	_ json.Unmarshaler = (*RedactedThinkingContent)(nil)
)

func (r *RedactedThinkingContent) isMessageContent() {}

type redactedThinkingContentJSON struct {
	Data string `json:"data"`
}

func (r RedactedThinkingContent) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]redactedThinkingContentJSON{
		"redacted_thinking": {Data: r.Data},
	})
}

func (r *RedactedThinkingContent) UnmarshalJSON(data []byte) error {
	var wrapper struct {
		RedactedThinking redactedThinkingContentJSON `json:"redacted_thinking"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return err
	}
	r.Data = wrapper.RedactedThinking.Data
	return nil
}
