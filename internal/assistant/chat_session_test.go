package assistant_test

import (
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"

	"micheam.com/aico/internal/assistant"
	assistantv1 "micheam.com/aico/internal/pb/assistant/v1"
)

func TestChatSession_Serialize(t *testing.T) {
	orig := &assistantv1.ChatSession{
		Id:        "sess-9D735EADE08A48E58C169E18E5AE3171",
		CreatedAt: timestamppb.Now(),
		SystemInstruction: &assistantv1.TextContent{
			Text: "This is a system instruction message.",
		},
		History: []*assistantv1.Message{
			{
				Role: assistantv1.Message_ROLE_USER,
				Contents: []*assistantv1.MessageContent{
					{
						Content: &assistantv1.MessageContent_Text{
							Text: &assistantv1.TextContent{
								Text: "Hello, how are you doing?",
							},
						},
					},
				},
			},
			{
				Role: assistantv1.Message_ROLE_ASSISTANT,
				Contents: []*assistantv1.MessageContent{
					{
						Content: &assistantv1.MessageContent_Text{
							Text: &assistantv1.TextContent{
								Text: "I am doing great, thank you!",
							},
						},
					},
				},
			},
		},
	}

	serialized, err := proto.Marshal(orig)
	if err != nil {
		t.Fatal(err)
	}

	var restored assistantv1.ChatSession
	if err := proto.Unmarshal(serialized, &restored); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(orig, &restored, protocmp.Transform()); diff != "" {
		t.Fatalf("unexpected diff (-want +got):\n%s", diff)
	}
}

func TestChatSession_ToMarkdown_LoadMarkdown_RoundTrip(t *testing.T) {
	// Create a test chat session with various content types
	originalSession := &assistant.ChatSession{
		ID:                "sess-test123456789",
		Title:             "Test Chat Session",
		CreatedAt:         time.Date(2025, 5, 29, 12, 0, 0, 0, time.UTC),
		SystemInstruction: assistant.NewTextContent("You are a helpful assistant.\nPlease be concise and clear."),
	}

	// Add some messages with different content types
	originalSession.History = []*assistant.Message{
		{
			Author: assistant.MessageAuthorUser,
			Contents: []assistant.MessageContent{
				assistant.NewTextContent("Hello, can you help me with Go programming?"),
			},
		},
		{
			Author: assistant.MessageAuthorAssistant,
			Contents: []assistant.MessageContent{
				assistant.NewTextContent("Sure! Here's a simple example:"),
				assistant.NewTextContent("```go\npackage main\n\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}\n```\n"),
			},
		},
		{
			Author: assistant.MessageAuthorUser,
			Contents: []assistant.MessageContent{
				assistant.NewTextContent("Can you also show me a file example?"),
				assistant.NewAttachmentContent("example.go", "go", "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Example\")\n}"),
			},
		},
		{
			Author: assistant.MessageAuthorAssistant,
			Contents: []assistant.MessageContent{
				assistant.NewTextContent("I can see your example file. It looks good!"),
			},
		},
	}

	// Convert to markdown
	markdown, err := originalSession.ToMarkdown()
	require.NoError(t, err, "ToMarkdown should not fail")

	// Verify markdown structure
	require.Contains(t, markdown, "---", "Markdown should contain frontmatter")
	require.Contains(t, markdown, "## System Instructions", "Markdown should contain System Instructions section")
	require.Contains(t, markdown, "## History", "Markdown should contain History section")
	require.Contains(t, markdown, "**User**", "Markdown should contain User messages")
	require.Contains(t, markdown, "**Assistant**", "Markdown should contain Assistant messages")
	require.Contains(t, markdown, "<details>", "Markdown should contain attachment details")

	// Load back from markdown
	restoredSession := &assistant.ChatSession{}
	err = assistant.LoadMarkdown(restoredSession, strings.NewReader(markdown))
	require.NoError(t, err, "LoadMarkdown should not fail")

	// Compare key fields
	require.Equal(t, originalSession.ID, restoredSession.ID, "Session ID should match")
	require.Equal(t, originalSession.Title, restoredSession.Title, "Session title should match")
	require.True(t, originalSession.CreatedAt.Equal(restoredSession.CreatedAt), "CreatedAt should match")
	require.Equal(t, originalSession.SystemInstruction.Text, restoredSession.SystemInstruction.Text, "System instruction should match")
	require.Equal(t, len(originalSession.History), len(restoredSession.History), "History length should match")

	// Compare each message
	for i, origMsg := range originalSession.History {
		require.Less(t, i, len(restoredSession.History), "Message %d should exist in restored session", i)

		restoredMsg := restoredSession.History[i]

		require.Equal(t, origMsg.Author, restoredMsg.Author, "Message %d author should match", i)
		require.Equal(t, len(origMsg.Contents), len(restoredMsg.Contents), "Message %d content count should match", i)

		// Compare content
		for j, origContent := range origMsg.Contents {
			restoredContent := restoredMsg.Contents[j]

			switch orig := origContent.(type) {
			case *assistant.TextContent:
				restored, ok := restoredContent.(*assistant.TextContent)
				require.True(t, ok, "Message %d content %d should be TextContent", i, j)
				require.Equal(t, orig.Text, restored.Text, "Message %d content %d text should match", i, j)
			case *assistant.AttachmentContent:
				restored, ok := restoredContent.(*assistant.AttachmentContent)
				require.True(t, ok, "Message %d content %d should be AttachmentContent", i, j)
				require.Equal(t, orig.Name, restored.Name, "Message %d attachment %d name should match", i, j)
				require.Equal(t, orig.Syntax, restored.Syntax, "Message %d attachment %d syntax should match", i, j)
				require.Equal(t, string(orig.Content), string(restored.Content), "Message %d attachment %d content should match", i, j)
			}
		}
	}
}
