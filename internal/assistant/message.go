package assistant

import (
	"fmt"
	"net/url"

	assistantpb "micheam.com/aico/internal/pb/assistant/v1"
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

func (m *Message) toProto() (*assistantpb.Message, error) {
	dest := &assistantpb.Message{}
	switch m.Author {
	case "assistant":
		dest.Role = assistantpb.Message_ROLE_ASSISTANT
	case "user":
		dest.Role = assistantpb.Message_ROLE_USER
	default:
		return nil, fmt.Errorf("unsupported message author: %q", m.Author)
	}
	for _, c := range m.Contents {
		switch c := c.(type) {
		case *TextContent:
			dest.Contents = append(dest.Contents, &assistantpb.MessageContent{
				Content: &assistantpb.MessageContent_Text{
					Text: &assistantpb.TextContent{
						Text: c.Text,
					},
				},
			})
		case *URLImageContent:
			dest.Contents = append(dest.Contents, &assistantpb.MessageContent{
				Content: &assistantpb.MessageContent_Image{
					Image: &assistantpb.URLImageContent{
						Url: c.URL.String(),
					},
				},
			})
		case *AttachmentContent:
			dest.Contents = append(dest.Contents, &assistantpb.MessageContent{
				Content: &assistantpb.MessageContent_Attachment{
					Attachment: &assistantpb.AttachmentContent{
						Name:    c.Name,
						Syntax:  c.Syntax,
						Content: c.Content,
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
