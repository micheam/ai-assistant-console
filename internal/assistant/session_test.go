package assistant

import (
	"testing"

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
