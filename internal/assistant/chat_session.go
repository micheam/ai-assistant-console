package assistant

import (
	"context"
	"fmt"
	"io"
	"iter"
	"net/url"
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
	ID                string          `json:"id"`
	SystemInstruction *TextContent    `json:"system_instruction"`
	History           []*Message      `json:"history"`
	Model             GenerativeModel `json:"-"`
	Title             string          `json:"title"`
	CreatedAt         time.Time       `json:"created_at"`
}

func (c *ChatSession) LastMessage() *Message {
	if len(c.History) == 0 {
		return nil
	}
	return c.History[len(c.History)-1]
}

// StartChat starts a new chat session.
func StartChat(m GenerativeModel) (*ChatSession, error) {
	newid, err := newChatSessionID()
	if err != nil {
		return nil, err
	}
	return &ChatSession{
		ID:        newid,
		Model:     m,
		CreatedAt: time.Now(),
	}, nil
}

// RestoreChat restores a chat session from the given ID.
func RestoreChat(dir, id string, m GenerativeModel) (*ChatSession, error) {
	filepath := filepath.Join(dir, id+".pb")
	f, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	sessPB := &assistantv1.ChatSession{}
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	err = proto.Unmarshal(b, sessPB)
	if err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	var sess = new(ChatSession)
	err = sess.fromProto(sessPB)
	if err != nil {
		return nil, fmt.Errorf("from proto: %w", err)
	}
	if m != nil {
		sess.Model = m
	}
	return sess, nil
}

func (c *ChatSession) SetSystemInstruction(text *TextContent) {
	c.SystemInstruction = text
}

func (c *ChatSession) GetSystemInstruction() *TextContent {
	if c.SystemInstruction == nil {
		return &TextContent{}
	}
	return c.SystemInstruction
}

// SendMessage sends a message to the chat session.
func (c *ChatSession) SendMessage(
	ctx context.Context,
	m ...MessageContent,
) (*GenerateContentResponse, error) {
	c.History = append(c.History, NewUserMessage(m...))
	c.Model.SetSystemInstruction(c.SystemInstruction)
	resp, err := c.Model.GenerateContent(ctx, c.History...)
	if err != nil {
		return nil, fmt.Errorf("generate content: %w", err)
	}
	c.AddHistory(resp)
	return resp, nil
}

// Continue continues the chat session with the last message.
func (c *ChatSession) Continue(ctx context.Context) (*GenerateContentResponse, error) {
	if len(c.History) == 0 {
		return nil, fmt.Errorf("no messages in history to continue")
	}
	lastMsg := c.LastMessage()
	if lastMsg.Author != MessageAuthorUser {
		return nil, fmt.Errorf("last message is not from user")
	}
	if c.Model == nil {
		return nil, fmt.Errorf("model is not set")
	}
	c.Model.SetSystemInstruction(c.GetSystemInstruction())
	resp, err := c.Model.GenerateContent(ctx, c.History...)
	if err != nil {
		return nil, fmt.Errorf("generate content: %w", err)
	}
	c.AddHistory(resp)
	return resp, nil
}

// ContinueStream continues the chat session with streaming for the last message.
func (c *ChatSession) ContinueStream(ctx context.Context) (
	iter.Seq[*GenerateContentResponse],
	error,
) {
	if len(c.History) == 0 {
		return nil, fmt.Errorf("no messages in history to continue")
	}
	lastMsg := c.LastMessage()
	if lastMsg.Author != MessageAuthorUser {
		return nil, fmt.Errorf("last message is not from user")
	}
	if c.Model == nil {
		return nil, fmt.Errorf("model is not set")
	}
	c.Model.SetSystemInstruction(c.GetSystemInstruction())
	iter, err := c.Model.GenerateContentStream(ctx, c.History...)
	if err != nil {
		return nil, fmt.Errorf("generate content stream: %w", err)
	}
	return iter, nil
}

// SendMessageStream sends a message to the chat session.
func (c *ChatSession) SendMessageStream(ctx context.Context, m ...MessageContent) (
	iter.Seq[*GenerateContentResponse],
	error,
) {
	c.History = append(c.History, NewUserMessage(m...))
	c.Model.SetSystemInstruction(c.SystemInstruction)
	iter, err := c.Model.GenerateContentStream(ctx, c.History...)
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

	filepath := filepath.Join(dir, c.ID+".pb")
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
		Id:        c.ID,
		CreatedAt: timestamppb.New(c.CreatedAt),
	}
	if c.SystemInstruction != nil {
		destPB.SystemInstruction = &assistantv1.TextContent{
			Text: c.SystemInstruction.Text,
		}
	}
	destPB.History = make([]*assistantv1.Message, 0, len(c.History))
	for _, msg := range c.History {
		msgPB, err := msg.toProto()
		if err != nil {
			return nil, fmt.Errorf("history.msg to proto: %w", err)
		}
		destPB.History = append(destPB.History, msgPB)
	}
	return destPB, nil
}

func (c *ChatSession) fromProto(src *assistantv1.ChatSession) error {
	c.ID = src.Id
	c.CreatedAt = src.CreatedAt.AsTime()
	if src.SystemInstruction != nil {
		c.SystemInstruction = NewTextContent(src.SystemInstruction.Text)
	}
	c.History = make([]*Message, 0, len(src.History))
	for _, msgPB := range src.History {
		var author MessageAuthor
		switch msgPB.Role {
		case assistantv1.Message_ROLE_USER:
			author = MessageAuthorUser
		case assistantv1.Message_ROLE_ASSISTANT:
			author = MessageAuthorAssistant
		default:
			return fmt.Errorf("unknown role: %v", msgPB.Role)
		}
		msg := &Message{
			Author:   author,
			Contents: []MessageContent{},
		}
		for _, contentPB := range msgPB.Contents {
			switch contentPB.Content.(type) {
			case *assistantv1.MessageContent_Text:
				c := NewTextContent(contentPB.GetText().Text)
				msg.Contents = append(msg.Contents, c)
			case *assistantv1.MessageContent_Image:
				url_, err := url.Parse(contentPB.GetImage().GetUrl())
				if err != nil {
					return fmt.Errorf("parse url: %w", err)
				}
				c := NewURLImageContent(*url_)
				msg.Contents = append(msg.Contents, c)
			case *assistantv1.MessageContent_Attachment:
				attachment := contentPB.GetAttachment()
				c := &AttachmentContent{
					Name:    attachment.Name,
					Syntax:  attachment.Syntax,
					Content: attachment.Content,
				}
				msg.Contents = append(msg.Contents, c)
			}
		}
		c.History = append(c.History, msg)
	}
	return nil
}

// AddHistory appends a response to the chat session history.
func (c *ChatSession) AddHistory(resp *GenerateContentResponse) {
	c.History = append(c.History, NewAssistantMessage(resp.Content))
}

// ToMarkdown converts the chat session to markdown format compatible with LoadMarkdown.
func (c *ChatSession) ToMarkdown() (string, error) {
	var sb strings.Builder
	
	// Write frontmatter
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("session_id: %s\n", c.ID))
	sb.WriteString(fmt.Sprintf("created_at: %s\n", c.CreatedAt.Format(time.RFC3339)))
	if c.Title != "" {
		sb.WriteString(fmt.Sprintf("title: \"%s\"\n", c.Title))
	}
	sb.WriteString("---\n\n")
	
	// Write system instructions
	sb.WriteString("## System Instructions\n\n")
	if c.SystemInstruction != nil && c.SystemInstruction.Text != "" {
		sb.WriteString(c.SystemInstruction.Text)
		if !strings.HasSuffix(c.SystemInstruction.Text, "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}
	
	// Write chat history
	sb.WriteString("## History\n\n")
	
	for i, msg := range c.History {
		// Write message header with bold formatting
		authorName := "**Assistant**"
		if msg.Author == MessageAuthorUser {
			authorName = "**User**"
		}
		sb.WriteString(fmt.Sprintf("### %d. %s\n", i+1, authorName))
		
		// Write message contents
		for _, content := range msg.Contents {
			switch c := content.(type) {
			case *TextContent:
				sb.WriteString(c.Text)
				if !strings.HasSuffix(c.Text, "\n") {
					sb.WriteString("\n")
				}
				sb.WriteString("\n")
			case *AttachmentContent:
				sb.WriteString("<details>\n\n")
				sb.WriteString(fmt.Sprintf("<summary>Attachment: %s</summary>\n\n", c.Name))
				sb.WriteString("```")
				if c.Syntax != "" {
					sb.WriteString(c.Syntax)
				}
				sb.WriteString("\n")
				sb.Write(c.Content)
				if !strings.HasSuffix(string(c.Content), "\n") {
					sb.WriteString("\n")
				}
				sb.WriteString("```\n\n")
				sb.WriteString("</details>\n\n")
			default:
				// For other content types, convert to string representation
				sb.WriteString(fmt.Sprintf("%v\n\n", content))
			}
		}
	}
	
	return sb.String(), nil
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
