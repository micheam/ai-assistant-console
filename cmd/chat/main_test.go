package main

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"micheam.com/aico/internal/openai"
)

func TestParseInputMessages(t *testing.T) {
	testcases := map[string]struct {
		input string
		want  []openai.Message
	}{
		"empty": {
			input: "",
			want:  []openai.Message{},
		},
		"without prompt": {
			input: `
Hello, I'm a human.
Nice to meet you.
How are you?`,
			want: []openai.Message{
				{
					Role:    openai.RoleUser,
					Content: "Hello, I'm a human.\nNice to meet you.\nHow are you?",
				},
			},
		},
		"separate with blank new-line": {
			input: `Hello, I'm a human.
Nice to meet you.

How are you?
`,
			want: []openai.Message{
				{
					Role:    openai.RoleUser,
					Content: "Hello, I'm a human.\nNice to meet you.",
				},
				{
					Role:    openai.RoleUser,
					Content: "How are you?",
				},
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
			want: []openai.Message{
				{
					Role:    openai.RoleUser,
					Content: "Hello, I'm a human.\nNice to meet you.",
				},
				{
					Role:    openai.RoleAssistant,
					Content: "Hello! Nice to meet you too. How can I assist you today?",
				},
				{
					Role:    openai.RoleUser,
					Content: "What is the weather today?",
				},
			},
		},
		"prefix with role": {
			input: `User: Hello, I'm a human.
Nice to meet you.

Assistant: Hello! Nice to meet you too. How can I assist you today?

User: What is the weather today?
`,
			want: []openai.Message{
				{
					Role:    openai.RoleUser,
					Content: "Hello, I'm a human.\nNice to meet you.",
				},
				{
					Role:    openai.RoleAssistant,
					Content: "Hello! Nice to meet you too. How can I assist you today?",
				},
				{
					Role:    openai.RoleUser,
					Content: "What is the weather today?",
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
