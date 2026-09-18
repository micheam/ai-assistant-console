package assistant

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/require"
)

var sessionJSONStr = `{
  "id": "session_12345",
  "system_instruction": [
	{"text": "You are a helpful assistant."},
  	{"text": "Provide concise answers to user queries."}
  ],
  "messages": [
    {
      "author": "user",
      "contents": [
        {"text": "Hello, how are you?"},
		{"url": "https://example.com/image.jpg"}
      ]
    },
    {
      "author": "assistant",
      "contents": [
        {"text": "I'm fine, thank you!"}
      ]
    }
  ]
}`

func TestSession_MarshalJSON(t *testing.T) {
	// Unmarshal from JSON
	sess := new(Session)
	err := sess.UnmarshalJSON([]byte(sessionJSONStr))
	require.NoError(t, err)
	msgs := sess.GetMessages()
	require.Len(t, msgs, 2)

	// Marshal back to JSON
	data, err := sess.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, sessionJSONStr, string(data))
}

func TestSession_UnmarshalJSON_WithoutSource_LeavesSourceNil(t *testing.T) {
	// sessionJSONStr predates the "source" field: old session files must
	// still decode cleanly, with Source left nil.
	sess := new(Session)
	require.NoError(t, sess.UnmarshalJSON([]byte(sessionJSONStr)))
	require.Nil(t, sess.Source)
}

func TestSession_SourceRoundTrip(t *testing.T) {
	sess := &Session{
		ID:     "with-source",
		Source: &SourceInfo{File: "/tmp/buffer.go", Name: "buffer.go"},
		Messages: []Message{
			NewUserMessage(NewTextContent("hello")),
		},
	}
	data, err := sess.MarshalJSON()
	require.NoError(t, err)
	require.Contains(t, string(data), `"source"`)

	got := new(Session)
	require.NoError(t, got.UnmarshalJSON(data))
	require.NotNil(t, got.Source)
	require.Equal(t, "/tmp/buffer.go", got.Source.File)
	require.Equal(t, "buffer.go", got.Source.Name)
}

func TestSourceInfo_String(t *testing.T) {
	tests := []struct {
		name string
		src  *SourceInfo
		want string
	}{
		{"nil", nil, ""},
		{"empty", &SourceInfo{}, ""},
		{"file only", &SourceInfo{File: "main.go"}, "main.go"},
		{"name only", &SourceInfo{Name: "buffer.go"}, "buffer.go"},
		{"file and name", &SourceInfo{File: "/tmp/buffer.go", Name: "buffer.go"}, "buffer.go (/tmp/buffer.go)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.src.String())
		})
	}
}

// writeSessionFile writes a Session to <dir>/<id>.json and returns its path.
func writeSessionFile(t *testing.T, dir string, sess *Session) string {
	t.Helper()
	data, err := sess.MarshalJSON()
	require.NoError(t, err)
	path := filepath.Join(dir, sess.ID+".json")
	require.NoError(t, os.WriteFile(path, data, 0o600))
	return path
}

func TestSessionPreview(t *testing.T) {
	t.Run("collapses newlines/tabs/runs of whitespace into single spaces", func(t *testing.T) {
		dir := t.TempDir()
		sess := &Session{
			ID: "collapse",
			Messages: []Message{
				NewUserMessage(NewTextContent("line1\n\n  line2\tend")),
				NewAssistantMessage(NewTextContent("reply")),
			},
		}
		path := writeSessionFile(t, dir, sess)

		preview, msgCount := sessionPreview(path)
		require.Equal(t, "line1 line2 end", preview)
		require.Equal(t, 2, msgCount)
	})

	t.Run("truncates without splitting a multibyte rune", func(t *testing.T) {
		dir := t.TempDir()
		text := strings.Repeat("あ", 100)
		sess := &Session{
			ID: "multibyte",
			Messages: []Message{
				NewUserMessage(NewTextContent(text)),
			},
		}
		path := writeSessionFile(t, dir, sess)

		preview, msgCount := sessionPreview(path)
		require.True(t, utf8.ValidString(preview))
		require.LessOrEqual(t, len(preview), 83) // 80 bytes + "..."
		require.Equal(t, 1, msgCount)
	})

	t.Run("skips <source> and <context> blocks", func(t *testing.T) {
		dir := t.TempDir()
		sess := &Session{
			ID: "skip-source",
			Messages: []Message{
				NewUserMessage(
					NewTextContent("<source>ignored</source>"),
					NewTextContent("actual prompt"),
				),
			},
		}
		path := writeSessionFile(t, dir, sess)

		preview, msgCount := sessionPreview(path)
		require.Equal(t, "actual prompt", preview)
		require.Equal(t, 1, msgCount)
	})

	t.Run("returns (empty) when there is no usable user text", func(t *testing.T) {
		dir := t.TempDir()
		sess := &Session{
			ID: "no-text",
			Messages: []Message{
				NewAssistantMessage(NewTextContent("assistant only")),
			},
		}
		path := writeSessionFile(t, dir, sess)

		preview, msgCount := sessionPreview(path)
		require.Equal(t, "(empty)", preview)
		require.Equal(t, 1, msgCount)
	})

	t.Run("returns empty result for a nonexistent path", func(t *testing.T) {
		preview, msgCount := sessionPreview(filepath.Join(t.TempDir(), "no-such-file.json"))
		require.Equal(t, "", preview)
		require.Equal(t, 0, msgCount)
	})
}
