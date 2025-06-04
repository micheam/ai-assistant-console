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

	// metadata
	require.Equal("Simple chat with code", sess.Title)
	require.Equal("2025-03-04T10:00:00Z", sess.CreatedAt.Format("2006-01-02T15:04:05Z"))

	// system instruction
	require.Equal("You are a helpful programming assistant.\n"+
		"Be concise and provide clear examples.",
		sess.GetSystemInstruction().Text)
	require.Equal(4, len(sess.History))

	// chat history (0)
	//
	// 	### 1. **User**
	// 	How do I create a simple HTTP server in Go?
	history1 := sess.History[0]
	require.Equal(assistant.MessageAuthorUser, history1.Author)
	require.Len(history1.GetContents(), 1)
	require.Equal("How do I create a simple HTTP server in Go?", history1.GetContents()[0].(*assistant.TextContent).Text)

	// chat history (1)
	//
	// 	### 2. **Assistant**
	// 	Here's a basic HTTP server in Go:
	//
	// 	```go
	// 	package main
	//
	// 	import (
	// 	    "fmt"
	// 	    "net/http"
	// 	)
	//
	// 	func handler(w http.ResponseWriter, r *http.Request) {
	// 	    fmt.Fprintf(w, "Hello, World!")
	// 	}
	//
	// 	func main() {
	// 	    http.HandleFunc("/", handler)
	// 	    http.ListenAndServe(":8080", nil)
	// 	}
	// 	```
	//
	// 	Run this and visit `http://localhost:8080` to see it in action.
	history2 := sess.History[1]
	require.Equal(assistant.MessageAuthorAssistant, history2.Author)
	require.Len(history2.GetContents(), 3)
	require.Equal("Here's a basic HTTP server in Go:", history2.GetContents()[0].(*assistant.TextContent).Text)
	require.Equal("```go\npackage main\n\nimport (\n\t\"fmt\"\n\t\"net/http\"\n)\n\nfunc handler(w http.ResponseWriter, r *http.Request) {\n\tfmt.Fprintf(w, \"Hello, World!\")\n}\n\nfunc main() {\n\thttp.HandleFunc(\"/\", handler)\n\thttp.ListenAndServe(\":8080\", nil)\n}\n```\n", history2.GetContents()[1].(*assistant.TextContent).Text)
	require.Equal("Run this and visit `http://localhost:8080` to see it in action.", history2.GetContents()[2].(*assistant.TextContent).Text)

	// chat history (2)
	//
	// 	### 3. **User**
	// 	Can you explain the code?
	history3 := sess.History[2]
	require.Equal(assistant.MessageAuthorUser, history3.Author)
	require.Len(history3.GetContents(), 1)
	require.Equal("Thanks!", history3.GetContents()[0].(*assistant.TextContent).Text)

	// chat history (3)
	//
	// 	### 4. **Assistant**
	// 	Of course! This code sets up a simple HTTP server that responds with "Hello, World!" when accessed.
	// 	The `handler` function is called whenever a request is made to the server, and it writes the response.
	// 	The `main` function initializes the server and listens on port 8080.
	// 	You can test it by running the code and visiting `http://localhost:8080` in your browser.
	history4 := sess.History[3]
	require.Equal(assistant.MessageAuthorAssistant, history4.Author)
	require.Len(history4.GetContents(), 1)
	require.Equal("You're welcome! If you have any more questions or need further assistance, feel free to ask.", history4.GetContents()[0].(*assistant.TextContent).Text)
}

func TestLoad_MessageWithArtifact(t *testing.T) {
	// Setup:
	require := require.New(t)
	file, err := os.Open("testdata/attachment.md")
	require.NoError(err)

	// Exercise:
	sess := &assistant.ChatSession{}
	err = assistant.LoadMarkdown(sess, file)
	require.NoError(err)

	// Verify:

	// User message:
	//
	// 	content[0]| Review the following code and provide feedback.
	//
	// 	content[1]| <details><summary>Artifact: test.go</summary>
	//            |
	// 	          | ```go
	// 	          | package main
	// 	          | import (
	// 	          |     "fmt"
	// 	          | )
	//            |
	// 	          | func main() {
	// 	          |     fmt.Println("Hello, World!")
	// 	          | }
	// 	          | ```
	//            |
	// 	          | </details>
	history1 := sess.History[0]
	require.Equal(assistant.MessageAuthorUser, history1.Author)

	require.Len(history1.GetContents(), 2)

	require.Equal(
		"Review the following code and provide feedback.",
		history1.GetContents()[0].(*assistant.TextContent).Text,
	)

	attachment := history1.GetContents()[1].(*assistant.AttachmentContent)
	require.Equal("test.go", attachment.Name)
	require.Equal("go", attachment.Syntax)
	require.Contains(string(attachment.Content), "package main")
	require.Contains(string(attachment.Content), "fmt.Println(\"Hello, World!\")")
}

func TestMatchesAssistant(t *testing.T) {
	require := require.New(t)

	// Test various Assistant header formats
	testCases := []struct {
		header   string
		expected bool
	}{
		{"Assistant", true},
		{"**Assistant**", true},
		{"1. Assistant", true},
		{"2. **Assistant**", true},
		{"User", false},
		{"**User**", false},
		{"Random text", false},
		{"assistant", true}, // case insensitive
		{"ASSISTANT", true}, // case insensitive
		{"*assistant*", true},
	}

	for _, tc := range testCases {
		t.Run(tc.header, func(t *testing.T) {
			result := assistant.MatchesAssistant(tc.header)
			require.Equal(tc.expected, result, "header: %s", tc.header)
		})
	}
}

func TestMatchesUser(t *testing.T) {
	require := require.New(t)

	// Test various User header formats
	testCases := []struct {
		header   string
		expected bool
	}{
		{"User", true},
		{"**User**", true},
		{"1. User", true},
		{"2. **User**", true},
		{"Assistant", false},
		{"**Assistant**", false},
		{"Random text", false},
		{"user", true}, // case insensitive
		{"USER", true}, // case insensitive
		{"*user*", true},
	}

	for _, tc := range testCases {
		t.Run(tc.header, func(t *testing.T) {
			result := assistant.MatchesUser(tc.header)
			require.Equal(tc.expected, result, "header: %s", tc.header)
		})
	}
}

func TestLoad_MessageHeaderVariations(t *testing.T) {
	// Setup:
	require := require.New(t)
	file, err := os.Open("testdata/assistant_headers_test.md")
	require.NoError(err)
	defer file.Close()

	// Exercise:
	sess := &assistant.ChatSession{}
	err = assistant.LoadMarkdown(sess, file)
	require.NoError(err)

	// Verify:
	require.Len(sess.History, 4)

	// Check that both User and Assistant headers with various formatting are properly recognized
	require.Equal(assistant.MessageAuthorUser, sess.History[0].Author)       // User
	require.Equal(assistant.MessageAuthorAssistant, sess.History[1].Author) // **Assistant**
	require.Equal(assistant.MessageAuthorUser, sess.History[2].Author)       // **User**
	require.Equal(assistant.MessageAuthorAssistant, sess.History[3].Author) // Assistant

	// Verify content
	require.Equal("Test message from user.", sess.History[0].GetContents()[0].(*assistant.TextContent).Text)
	require.Equal("Response from assistant with bold formatting.", sess.History[1].GetContents()[0].(*assistant.TextContent).Text)
	require.Equal("Another user message with bold formatting.", sess.History[2].GetContents()[0].(*assistant.TextContent).Text)
	require.Equal("Response from assistant without formatting.", sess.History[3].GetContents()[0].(*assistant.TextContent).Text)
}
