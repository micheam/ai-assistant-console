package assistant

import (
	"context"
	"fmt"
	"iter"
	"time"

	"github.com/google/uuid"
)

// ChatSession represents a chat session with the assistant.
type ChatSession struct {
	id                string
	systemInstruction *TextContent
	history           []*Message

	model GenerativeModel

	title     string
	createdAt time.Time
}

// StartChat starts a new chat session.
func StartChat(m GenerativeModel) (*ChatSession, error) {
	newid, err := newChatSessionID()
	if err != nil {
		return nil, err
	}
	return &ChatSession{
		id:        newid,
		model:     m,
		createdAt: time.Now(),
	}, nil
}

func (c *ChatSession) SetSystemInstruction(text *TextContent) {
	c.systemInstruction = text
}

// SendMessage sends a message to the chat session.
func (c *ChatSession) SendMessage(
	ctx context.Context,
	m ...MessageContent,
) (*GenerateContentResponse, error) {
	c.history = append(c.history, NewUserMessage(m...))
	c.model.SetSystemInstruction(c.systemInstruction)
	resp, err := c.model.GenerateContent(ctx, c.history...)
	if err != nil {
		return nil, fmt.Errorf("generate content: %w", err)
	}
	c.addHistory(resp)
	return resp, nil
}

// SendMessageStream sends a message to the chat session.
func (c *ChatSession) SendMessageStream(
	ctx context.Context,
	m ...MessageContent,
) (iter.Seq[*GenerateContentResponse], error) {
	c.history = append(c.history, NewUserMessage(m...))
	c.model.SetSystemInstruction(c.systemInstruction)
	iter, err := c.model.GenerateContentStream(ctx, c.history...)
	if err != nil {
		return nil, fmt.Errorf("generate content stream: %w", err)
	}
	return iter, nil
}

// addHistory adds a response to the chat session history.
func (c *ChatSession) addHistory(resp *GenerateContentResponse) {
	c.history = append(c.history, NewAssistantMessage(resp.Content))
}

// newChatSessionID generates a new chat session ID.
func newChatSessionID() (string, error) {
	rawID, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return rawID.String(), nil
}

// GenerativeModel represents a generative model.
type GenerativeModel interface {
	Description() string
	SetSystemInstruction(*TextContent)
	GenerateContent(context.Context, ...*Message) (*GenerateContentResponse, error)
	GenerateContentStream(context.Context, ...*Message) (iter.Seq[*GenerateContentResponse], error)
}

type GenerateContentResponse struct {
	Content MessageContent
}
