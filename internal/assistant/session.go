package assistant

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"micheam.com/aico/internal/logging"
)

type Session struct {
	ID                string         `json:"id"`
	Model             string         `json:"model,omitempty"`
	SystemInstruction []*TextContent `json:"system_instruction"`
	Messages          []Message      `json:"messages"`

	filePath string `json:"-"`
}

func (s Session) GetMessages() []Message {
	msg := make([]Message, len(s.Messages))
	copy(msg, s.Messages)
	return msg
}

func (s *Session) AddMessage(message Message) {
	s.Messages = append(s.Messages, message)
}

func (s *Session) AddMessages(messages ...Message) {
	s.Messages = append(s.Messages, messages...)
}

// NewSession creates a new session with a unique ID and optional initial messages.
// If no messages are provided, the session will start empty.
//
// Note: Loading an existing session should be done via [LoadSession].
func NewSession(dir string, messages ...Message) *Session {
	id := uuid.NewString()
	return &Session{
		ID:       id,
		Messages: messages,
		filePath: sessionFilePath(dir, id),
	}
}

// -------------------------------------------
// Helper: Load/Save Session
// -------------------------------------------

func sessionFilePath(dir, id string) string {
	return filepath.Join(dir, id+".json")
}

func (s *Session) FilePath() string {
	return s.filePath
}

func LoadSession(ctx context.Context, dir, id string) (*Session, error) {
	f, err := os.Open(sessionFilePath(dir, id))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	sess, err := decodeSession(f)
	if err != nil {
		return nil, err
	}
	// Set Unexported fields...
	sess.filePath = sessionFilePath(dir, id)
	return sess, nil
}

// Save saves the session to file.
// This will overwrite any existing session file.
func (s *Session) Save(ctx context.Context, model GenerativeModel) error {
	logger := logging.LoggerFrom(ctx)

	dir := filepath.Dir(s.FilePath())
	stat, err := os.Stat(dir)
	if err != nil || !stat.IsDir() {
		if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
			return fmt.Errorf("failed to create session directory: %w", mkErr)
		}
	}
	f, err := os.OpenFile(s.FilePath(), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open session file for writing: %w", err)
	}
	defer f.Close()
	logger.Debug("saving session", "file", s.FilePath(), "model", model.Name())
	return s.encode(f)
}

// -------------------------------------------
// Helper: Encoding/Decoding Session into file
// -------------------------------------------

func decodeSession(file *os.File) (*Session, error) {
	var sess Session
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&sess); err != nil {
		return nil, err
	}
	return &sess, nil
}

func (s *Session) encode(file *os.File) error {
	encoder := json.NewEncoder(file)
	return encoder.Encode(s)
}

var (
	_ json.Marshaler   = (*Session)(nil)
	_ json.Unmarshaler = (*Session)(nil)
)

func (s Session) MarshalJSON() ([]byte, error) {
	type alias Session
	return json.Marshal(&struct {
		*alias
	}{
		alias: (*alias)(&s),
	})
}

func (s *Session) UnmarshalJSON(data []byte) error {
	var temp struct {
		ID                string            `json:"id"`
		Model             string            `json:"model,omitempty"`
		SystemInstruction []json.RawMessage `json:"system_instruction"`
		Messages          []json.RawMessage `json:"messages"`
	}

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	s.ID = temp.ID
	s.Model = temp.Model

	// Unmarshal system instructions
	s.SystemInstruction = make([]*TextContent, 0, len(temp.SystemInstruction))
	for _, raw := range temp.SystemInstruction {
		var tc TextContent
		if err := json.Unmarshal(raw, &tc); err != nil {
			return err
		}
		s.SystemInstruction = append(s.SystemInstruction, &tc)
	}

	// Unmarshal messages by detecting their type
	s.Messages = make([]Message, 0, len(temp.Messages))
	for _, raw := range temp.Messages {
		// Determine message type by looking at the author field
		var msgType struct {
			Author MessageAuthor `json:"author"`
		}
		if err := json.Unmarshal(raw, &msgType); err != nil {
			return err
		}

		var msg Message
		switch msgType.Author {
		case MessageAuthorUser:
			var um UserMessage
			if err := json.Unmarshal(raw, &um); err != nil {
				return err
			}
			msg = &um
		case MessageAuthorAssistant:
			var am AssistantMessage
			if err := json.Unmarshal(raw, &am); err != nil {
				return err
			}
			msg = &am
		}

		if msg != nil {
			s.Messages = append(s.Messages, msg)
		}
	}

	return nil
}

// -------------------------------------------
// Session Listing
// -------------------------------------------

// SessionSummary holds lightweight metadata for session listing.
type SessionSummary struct {
	ID       string
	ModTime  time.Time
	Preview  string // first user message preview
	MsgCount int
}

// ListSessions returns summaries of all sessions in the given directory,
// sorted by modification time (newest first).
func ListSessions(dir string) ([]SessionSummary, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read session dir: %w", err)
	}

	var summaries []SessionSummary
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".json")
		preview, msgCount := sessionPreview(filepath.Join(dir, e.Name()))
		summaries = append(summaries, SessionSummary{
			ID:       id,
			ModTime:  info.ModTime(),
			Preview:  preview,
			MsgCount: msgCount,
		})
	}

	slices.SortFunc(summaries, func(a, b SessionSummary) int {
		return b.ModTime.Compare(a.ModTime) // newest first
	})
	return summaries, nil
}

// LatestSessionID returns the ID of the most recently modified session.
func LatestSessionID(dir string) (string, error) {
	summaries, err := ListSessions(dir)
	if err != nil {
		return "", err
	}
	if len(summaries) == 0 {
		return "", fmt.Errorf("no sessions found")
	}
	return summaries[0].ID, nil
}

// sessionPreview reads a session file and extracts the first user message text.
func sessionPreview(path string) (string, int) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0
	}
	defer f.Close()

	sess, err := decodeSession(f)
	if err != nil {
		return "", 0
	}

	for _, msg := range sess.Messages {
		if msg.GetAuthor() != MessageAuthorUser {
			continue
		}
		for _, c := range msg.GetContents() {
			tc, ok := c.(*TextContent)
			if !ok || tc.Text == "" {
				continue
			}
			text := tc.Text
			if strings.HasPrefix(text, "<source") || strings.HasPrefix(text, "<context") {
				continue
			}
			if len(text) > 80 {
				text = text[:80] + "..."
			}
			return text, len(sess.Messages)
		}
	}
	return "(empty)", len(sess.Messages)
}
