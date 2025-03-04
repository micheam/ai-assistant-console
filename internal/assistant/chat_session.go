package assistant

// TODO: Rename file to chat_session.go

import (
	"context"
	"fmt"
	"iter"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	assistantv1 "micheam.com/aico/internal/pb/assistant/v1"
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

// Save saves the chat session.
//
// Note: This method will create the directory if it does not exist.
func (c *ChatSession) Save(dir string) error {
	// mkdir if not exists
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	// save the session
	sessPB, err := c.toProto()
	if err != nil {
		return fmt.Errorf("to proto: %w", err)
	}
	serialized, err := proto.Marshal(sessPB)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	filepath := filepath.Join(dir, c.id+".pb")
	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()
	_, err = f.Write(serialized)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

func (c *ChatSession) toProto() (*assistantv1.ChatSession, error) {
	destPB := &assistantv1.ChatSession{
		Id:        c.id,
		CreatedAt: timestamppb.New(c.createdAt),
	}
	if c.systemInstruction != nil {
		destPB.SystemInstruction = &assistantv1.TextContent{
			Text: c.systemInstruction.Text,
		}
	}
	destPB.History = make([]*assistantv1.Message, 0, len(c.history))
	for _, msg := range c.history {
		msgPB, err := msg.toProto()
		if err != nil {
			return nil, fmt.Errorf("history.msg to proto: %w", err)
		}
		destPB.History = append(destPB.History, msgPB)
	}
	return destPB, nil
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
	id := "sess-" + strings.ReplaceAll(rawID.String(), "-", "")
	return id, nil
}
