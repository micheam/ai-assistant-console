package main

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"micheam.com/aico/internal/chat"
)

func TestParseInputMessages(t *testing.T) {
	testcases := map[string]struct {
		input string
		want  []chat.Message
	}{
		"empty": {
			input: "",
			want:  []chat.Message{},
		},
		"without prompt": {
			input: `Hello, I'm a human.
Nice to meet you.
How are you?`,
			want: []chat.Message{
				&chat.UserMessage{
					Content: []chat.Content{&chat.TextContent{Text: "Hello, I'm a human.\nNice to meet you.\nHow are you?"}},
				},
			},
		},
		"separate with blank new-line": {
			input: `Hello, I'm a human.
Nice to meet you.

How are you?
`,
			want: []chat.Message{
				&chat.UserMessage{Content: []chat.Content{&chat.TextContent{Text: "Hello, I'm a human.\nNice to meet you."}}},
				&chat.UserMessage{Content: []chat.Content{&chat.TextContent{Text: "How are you?"}}},
			},
		},
		"prompt line with role": {
			input: `User:
Hello, I'm a human.
Nice to meet you.

Assistant:
Hello! Nice to meet you too. How can I assist you today?

User:
What is the weather today?
`,
			want: []chat.Message{
				&chat.UserMessage{
					Content: []chat.Content{&chat.TextContent{Text: "Hello, I'm a human.\nNice to meet you."}},
				},
				&chat.AssistantMessage{
					Content: []chat.Content{&chat.TextContent{Text: "Hello! Nice to meet you too. How can I assist you today?"}},
				},
				&chat.UserMessage{
					Content: []chat.Content{&chat.TextContent{Text: "What is the weather today?"}},
				},
			},
		},
		"trail characters after prompt will be ignored": {
			input: `User: --------------------------------
Hello, I'm a human.
Nice to meet you.

Assistant: -----------------------------------
Hello! Nice to meet you too. How can I assist you today?

User: --------------------------------
What is the weather today?
`,
			want: []chat.Message{
				&chat.UserMessage{
					Content: []chat.Content{&chat.TextContent{Text: "Hello, I'm a human.\nNice to meet you."}},
				},
				&chat.AssistantMessage{
					Content: []chat.Content{&chat.TextContent{Text: "Hello! Nice to meet you too. How can I assist you today?"}},
				},
				&chat.UserMessage{
					Content: []chat.Content{&chat.TextContent{Text: "What is the weather today?"}},
				},
			},
		},
	}
	for name, tc := range testcases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			r := strings.NewReader(tc.input)
			got := ParseInputMessage(r)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("parseInputMessages() mismatch (-want +got):\n%s", diff)
			}
		})
	}

}
