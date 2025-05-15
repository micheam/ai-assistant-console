package assistant_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"micheam.com/aico/internal/assistant"
)

func TestLoad_SimpleChat(t *testing.T) {
	// Setup:
	require := require.New(t)
	file, err := os.Open("testdata/simple_chat_history.md")
	require.NoError(err)

	// Exercise:
	sess := &assistant.ChatSession{}
	err = assistant.LoadMarkdown(sess, file)
	require.NoError(err)

	// Verify:
	require.Equal("Simple chat with code", sess.Title)
	require.Equal("2025-03-04T10:00:00Z", sess.CreatedAt.Format("2006-01-02T15:04:05Z"))
	require.Equal(4, len(sess.History))
}
