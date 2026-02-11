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
		var m map[string]interface{}
		if err := json.Unmarshal(raw, &m); err != nil {
			return err
		}

		var content MessageContent
		if _, hasText := m["text"]; hasText {
			var tc TextContent
			if err := json.Unmarshal(raw, &tc); err != nil {
				return err
			}
			content = &tc
		} else if _, hasURL := m["url"]; hasURL {
			var uc URLImageContent
			if err := json.Unmarshal(raw, &uc); err != nil {
				return err
			}
			content = &uc
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
		var m map[string]interface{}
		if err := json.Unmarshal(raw, &m); err != nil {
			return err
		}

		var content MessageContent
		if _, hasText := m["text"]; hasText {
			var tc TextContent
			if err := json.Unmarshal(raw, &tc); err != nil {
				return err
			}
			content = &tc
		} else if _, hasURL := m["url"]; hasURL {
			var uc URLImageContent
			if err := json.Unmarshal(raw, &uc); err != nil {
				return err
			}
			content = &uc
		}
		if content != nil {
			a.Contents = append(a.Contents, content)
		}
	}

	return nil
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
