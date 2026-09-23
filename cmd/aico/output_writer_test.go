package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"micheam.com/aico/internal/assistant"
)

func TestJSONLineStreamWriter_EmitToolUse_FlushesPendingPartialLineFirst(t *testing.T) {
	var buf bytes.Buffer
	w := &JSONLineStreamWriter{
		enc:      json.NewEncoder(&buf),
		metaData: JSONOutputMetaData{Session: "sess1", Model: "anthropic:claude-fable-5-1"},
	}

	_, err := w.Write([]byte("abc"))
	require.NoError(t, err)

	tu := &assistant.ToolUseContent{
		ID:    "toolu_01",
		Name:  "propose_edit",
		Input: json.RawMessage(`{"description":"rename x","old_string":"x","new_string":"y"}`),
	}
	require.NoError(t, w.EmitToolUse(tu))
	require.NoError(t, w.Close())

	// Close emits nothing further: EmitToolUse already flushed and reset the
	// buffer, and no usage was set.
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	require.Len(t, lines, 2) // flushed "abc", then the tool_use record

	var first jsonlModel
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &first))
	require.Equal(t, "abc", first.Content)
	require.Nil(t, first.ToolUse)

	var second jsonlModel
	require.NoError(t, json.Unmarshal([]byte(lines[1]), &second))
	require.Equal(t, "", second.Content)
	require.NotNil(t, second.ToolUse)
	require.Equal(t, "toolu_01", second.ToolUse.ID)
	require.Equal(t, "propose_edit", second.ToolUse.Name)
	require.JSONEq(t, string(tu.Input), string(second.ToolUse.Input))
}

func TestJSONLineStreamWriter_Close_EmitsUsage(t *testing.T) {
	var buf bytes.Buffer
	w := &JSONLineStreamWriter{
		enc:      json.NewEncoder(&buf),
		metaData: JSONOutputMetaData{Session: "sess1", Model: "m"},
	}
	w.SetUsage(&assistant.Usage{InputTokens: 10, OutputTokens: 5})
	require.NoError(t, w.Close())

	var got jsonlModel
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))
	require.NotNil(t, got.Usage)
	require.Equal(t, 10, got.Usage.InputTokens)
}

func TestConsoleLineStreamWriter_EmitToolUse_FlushesPendingPartialLine(t *testing.T) {
	var buf bytes.Buffer
	w := &ConsoleLineStreamWriter{out: &buf}

	_, err := w.Write([]byte("partial"))
	require.NoError(t, err)

	tu := &assistant.ToolUseContent{
		ID:    "toolu_01",
		Name:  "propose_edit",
		Input: json.RawMessage(`{"description":"rename x to y","old_string":"x = 1","new_string":"y = 1"}`),
	}
	require.NoError(t, w.EmitToolUse(tu))

	out := buf.String()
	require.Contains(t, out, "partial")
	require.Contains(t, out, "[propose_edit] rename x to y")
	require.Contains(t, out, "-x = 1")
	require.Contains(t, out, "+y = 1")
}
