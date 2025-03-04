package assistant

import (
	"fmt"
	"net/url"

	assistantv1 "micheam.com/aico/internal/pb/assistant/v1"
)

type Message struct {
	// Author is the author of the message.
	// e.g: "assistant", "user"
	Author string

	// Contents holds the contents of the message.
	Contents []MessageContent
}

func (m *Message) toProto() (*assistantv1.Message, error) {
	dest := &assistantv1.Message{}
	switch m.Author {
	case "assistant":
		dest.Role = assistantv1.Message_ROLE_ASSISTANT
	case "user":
		dest.Role = assistantv1.Message_ROLE_USER
	default:
		return nil, fmt.Errorf("unsupported message author: %s", m.Author)
	}
	for _, c := range m.Contents {
		switch c := c.(type) {
		case *TextContent:
			dest.Contents = append(dest.Contents, &assistantv1.MessageContent{
				Content: &assistantv1.MessageContent_Text{
					Text: &assistantv1.TextContent{
						Text: c.Text,
					},
				},
			})
		case *URLImageContent:
			dest.Contents = append(dest.Contents, &assistantv1.MessageContent{
				Content: &assistantv1.MessageContent_Image{
					Image: &assistantv1.URLImageContent{
						Url: c.URL.String(),
					},
				},
			})
		default:
			return nil, fmt.Errorf("unsupported message content type: %T", c)
		}
	}
	return dest, nil
}

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
	Text string
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
	URL url.URL
}

var _ MessageContent = (*URLImageContent)(nil)

func (u *URLImageContent) isMessageContent() {}

func NewURLImageContent(url url.URL) *URLImageContent {
	return &URLImageContent{URL: url}
}
