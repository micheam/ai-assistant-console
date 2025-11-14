package assistant

import (
	"net/url"
)

type Message struct {
	// Author is the author of the message.
	// e.g: "assistant", "user"
	Author MessageAuthor

	// Contents holds the contents of the message.
	Contents []MessageContent
}

func (m Message) GetContents() []MessageContent {
	if m.Contents == nil {
		return []MessageContent{}
	}
	return m.Contents
}

type MessageAuthor string

const (
	MessageAuthorAssistant MessageAuthor = "assistant"
	MessageAuthorUser      MessageAuthor = "user"
)

type MessageContent interface {
	isMessageContent()
}

// NewUserMessage creates a new user message.
func NewUserMessage(contents ...MessageContent) *Message {
	return &Message{
		Author:   "user",
		Contents: contents,
	}
}

// NewAssistantMessage creates a new assistant message.
func NewAssistantMessage(contents ...MessageContent) *Message {
	return &Message{
		Author:   "assistant",
		Contents: contents,
	}
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

var _ MessageContent = (*TextContent)(nil)

func (t *TextContent) isMessageContent() {}

func NewTextContent(text string) *TextContent {
	return &TextContent{Text: text}
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

var _ MessageContent = (*URLImageContent)(nil)

func (u *URLImageContent) isMessageContent() {}

func NewURLImageContent(url url.URL) *URLImageContent {
	return &URLImageContent{URL: url}
}
