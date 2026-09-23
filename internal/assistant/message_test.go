package assistant

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToolUseContent_RoundTrip(t *testing.T) {
	msg := NewAssistantMessage(
		NewTextContent("here is a fix"),
		&ToolUseContent{ID: "toolu_01", Name: "propose_edit", Input: json.RawMessage(`{"description":"rename x","old_string":"x","new_string":"y"}`)},
	)
	data, err := msg.MarshalJSON()
	require.NoError(t, err)
	require.Contains(t, string(data), `"tool_use"`)

	got := new(AssistantMessage)
	require.NoError(t, got.UnmarshalJSON(data))
	require.Len(t, got.Contents, 2)

	tu, ok := got.Contents[1].(*ToolUseContent)
	require.True(t, ok)
	require.Equal(t, "toolu_01", tu.ID)
	require.Equal(t, "propose_edit", tu.Name)
	require.JSONEq(t, `{"description":"rename x","old_string":"x","new_string":"y"}`, string(tu.Input))
}

func TestToolResultContent_RoundTrip(t *testing.T) {
	msg := NewUserMessage(
		&ToolResultContent{ToolUseID: "toolu_01", Content: "applied", IsError: false},
		NewTextContent("thanks"),
	)
	data, err := msg.MarshalJSON()
	require.NoError(t, err)

	got := new(UserMessage)
	require.NoError(t, got.UnmarshalJSON(data))
	require.Len(t, got.Contents, 2)

	tr, ok := got.Contents[0].(*ToolResultContent)
	require.True(t, ok)
	require.Equal(t, "toolu_01", tr.ToolUseID)
	require.Equal(t, "applied", tr.Content)
	require.False(t, tr.IsError)
}

func TestToolResultContent_IsError(t *testing.T) {
	msg := NewUserMessage(&ToolResultContent{ToolUseID: "toolu_02", Content: "not found", IsError: true})
	data, err := msg.MarshalJSON()
	require.NoError(t, err)

	got := new(UserMessage)
	require.NoError(t, got.UnmarshalJSON(data))
	tr := got.Contents[0].(*ToolResultContent)
	require.True(t, tr.IsError)
}

func TestThinkingContent_RoundTrip(t *testing.T) {
	msg := NewAssistantMessage(
		&ThinkingContent{Thinking: "", Signature: "Er1234=="},
		NewTextContent("answer"),
	)
	data, err := msg.MarshalJSON()
	require.NoError(t, err)

	got := new(AssistantMessage)
	require.NoError(t, got.UnmarshalJSON(data))
	require.Len(t, got.Contents, 2)

	th, ok := got.Contents[0].(*ThinkingContent)
	require.True(t, ok)
	require.Equal(t, "", th.Thinking)
	require.Equal(t, "Er1234==", th.Signature)
}

func TestRedactedThinkingContent_RoundTrip(t *testing.T) {
	msg := NewAssistantMessage(&RedactedThinkingContent{Data: "opaque-data"})
	data, err := msg.MarshalJSON()
	require.NoError(t, err)

	got := new(AssistantMessage)
	require.NoError(t, got.UnmarshalJSON(data))
	require.Len(t, got.Contents, 1)

	rt, ok := got.Contents[0].(*RedactedThinkingContent)
	require.True(t, ok)
	require.Equal(t, "opaque-data", rt.Data)
}

func TestDecodeContent_UnknownKey_ReturnsNilWithoutError(t *testing.T) {
	msg := new(UserMessage)
	raw := `{"author":"user","contents":[{"future_type":{"x":1}},{"text":"hello"}]}`
	require.NoError(t, msg.UnmarshalJSON([]byte(raw)))
	require.Len(t, msg.Contents, 1)
	tc, ok := msg.Contents[0].(*TextContent)
	require.True(t, ok)
	require.Equal(t, "hello", tc.Text)
}
