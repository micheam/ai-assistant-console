package openai

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"micheam.com/aico/internal/pointer"
)

func TestSystemMessage_MarshalJSON(t *testing.T) {
	type fields struct {
		Content string
		Name    *string
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]any
		wantErr bool
	}{
		{
			name: "marshal system message",
			want: map[string]any{
				"role":    "system",
				"content": "You are a helpful assistant.",
			},
			fields: fields{
				Content: "You are a helpful assistant.",
			},
		},
		{
			name: "marshal system message with name",
			want: map[string]any{
				"role":    "system",
				"content": "You are a helpful assistant.",
				"name":    "assistant",
			},
			fields: fields{
				Content: "You are a helpful assistant.",
				Name:    pointer.Ptr("assistant"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := SystemMessage{
				Content: tt.fields.Content,
				Name:    tt.fields.Name,
			}
			gotBytes, err := msg.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("SystemMessage.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var got map[string]interface{}
			err = json.Unmarshal(gotBytes, &got)
			require.NoError(t, err, "Failed to unmarshal JSON")

			require.Equal(t, tt.want, got, "SystemMessage.MarshalJSON() = %v, want %v", got, tt.want)
		})
	}
}

func TestSystemMessage_UnmarshalJSON(t *testing.T) {
	type fields struct {
		Content string
		Name    *string
	}
	tests := []struct {
		name    string
		fields  fields
		args    []byte
		want    SystemMessage
		wantErr bool
	}{
		{
			name: "unmarshal system message",
			args: []byte(`{"role":"system","content":"You are a helpful assistant."}`),
			want: SystemMessage{
				Content: "You are a helpful assistant.",
			},
		},
		{
			name: "unmarshal system message with name",
			args: []byte(`{"role":"system","content":"You are a helpful assistant.","name":"assistant"}`),
			want: SystemMessage{
				Content: "You are a helpful assistant.",
				Name:    pointer.Ptr("assistant"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg SystemMessage
			if err := msg.UnmarshalJSON(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("SystemMessage.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.Equal(t, tt.want, msg, "SystemMessage.UnmarshalJSON() = %v, want %v", msg, tt.want)
		})
	}
}

func TestUserMessage_MarshalJSON(t *testing.T) {
	imageURL := must(url.Parse("https://example.com/image.jpg"))
	input := UserMessage{
		Content: []Content{
			&TextContent{Text: "Hello, how are you?"},
			&ImageContent{URL: *imageURL},
		},
		Name: pointer.Ptr("user1"),
	}

	gotByte, err := input.MarshalJSON()
	require.NoError(t, err)

	// Unmarshal the JSON back to UserMessage
	var output UserMessage
	err = output.UnmarshalJSON(gotByte)
	require.NoError(t, err)

	// Compare the input and output
	if diff := cmp.Diff(input, output); diff != "" {
		t.Errorf("UserMessage.MarshalJSON() mismatch (-want +got):\n%s", diff)
	}
}

func TestAssistantMessage_MarshalJSON(t *testing.T) {
	type fields struct {
		Content []Content
		Name    *string
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]any
		wantErr bool
	}{
		{
			name: "marshal assistant message",
			want: map[string]any{
				"role":    "assistant",
				"content": []Content{&TextContent{Text: "How can I assist you?"}},
			},
			fields: fields{
				Content: []Content{&TextContent{Text: "How can I assist you?"}},
			},
		},
		{
			name: "marshal assistant message with name",
			want: map[string]any{
				"role":    "assistant",
				"content": []Content{&TextContent{Text: "How can I assist you?"}},
				"name":    "assistant1",
			},
			fields: fields{
				Content: []Content{&TextContent{Text: "How can I assist you?"}},
				Name:    pointer.Ptr("assistant1"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := AssistantMessage{
				Content: tt.fields.Content,
				Name:    tt.fields.Name,
			}
			gotBytes, err := msg.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("AssistantMessage.MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var got map[string]interface{}
			err = json.Unmarshal(gotBytes, &got)
			require.NoError(t, err, "Failed to unmarshal JSON")
			require.Equal(t, tt.want, got, "AssistantMessage.MarshalJSON() = %v, want %v", got, tt.want)
		})
	}
}

func TestAssistantMessage_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		args    []byte
		want    AssistantMessage
		wantErr bool
	}{
		{
			name: "unmarshal assistant message",
			args: []byte(`{"role":"assistant","content":[{"type":"text","content":"How can I assist you?"}]}`),
			want: AssistantMessage{
				Content: []Content{&TextContent{Text: "How can I assist you?"}},
			},
		},
		{
			name: "unmarshal assistant message with name",
			args: []byte(`{"role":"assistant","content":[{"type":"text","content":"How can I assist you?"}],"name":"assistant1"}`),
			want: AssistantMessage{
				Content: []Content{&TextContent{Text: "How can I assist you?"}},
				Name:    pointer.Ptr("assistant1"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var msg AssistantMessage
			if err := msg.UnmarshalJSON(tt.args); (err != nil) != tt.wantErr {
				t.Errorf("AssistantMessage.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.Equal(t, tt.want, msg, "AssistantMessage.UnmarshalJSON() = %v, want %v", msg, tt.want)
		})
	}
}

// Helper functions

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
