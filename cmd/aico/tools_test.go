package main

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"micheam.com/aico/internal/assistant"
)

func TestResolveTools(t *testing.T) {
	t.Run("resolves a known name", func(t *testing.T) {
		defs, err := resolveTools([]string{"propose_edit"})
		require.NoError(t, err)
		require.Len(t, defs, 1)
		require.Equal(t, "propose_edit", defs[0].Name)
	})

	t.Run("dedupes and sorts", func(t *testing.T) {
		defs, err := resolveTools([]string{"propose_edit", "propose_edit", ""})
		require.NoError(t, err)
		require.Len(t, defs, 1)
	})

	t.Run("errors on an unknown name", func(t *testing.T) {
		_, err := resolveTools([]string{"bogus"})
		require.Error(t, err)
		require.Contains(t, err.Error(), "bogus")
	})

	t.Run("empty input returns no tools", func(t *testing.T) {
		defs, err := resolveTools(nil)
		require.NoError(t, err)
		require.Empty(t, defs)
	})
}

func TestUnionToolNames(t *testing.T) {
	t.Run("merges and dedupes", func(t *testing.T) {
		got := unionToolNames([]string{"propose_edit"}, []string{"propose_edit", "other"})
		require.Equal(t, []string{"other", "propose_edit"}, got)
	})

	t.Run("both empty", func(t *testing.T) {
		require.Empty(t, unionToolNames(nil, nil))
	})
}

func TestValidateProposeEditInput(t *testing.T) {
	valid := json.RawMessage(`{"description":"rename x to y","old_string":"x = 1","new_string":"y = 1"}`)

	t.Run("valid input passes", func(t *testing.T) {
		require.NoError(t, validateProposeEditInput(valid))
	})

	t.Run("invalid JSON", func(t *testing.T) {
		err := validateProposeEditInput(json.RawMessage(`{not json`))
		require.Error(t, err)
	})

	t.Run("missing field", func(t *testing.T) {
		err := validateProposeEditInput(json.RawMessage(`{"description":"x","old_string":"a"}`))
		require.Error(t, err)
		require.Contains(t, err.Error(), "new_string")
	})

	t.Run("non-string field", func(t *testing.T) {
		err := validateProposeEditInput(json.RawMessage(`{"description":"x","old_string":1,"new_string":"a"}`))
		require.Error(t, err)
	})

	t.Run("empty old_string", func(t *testing.T) {
		err := validateProposeEditInput(json.RawMessage(`{"description":"x","old_string":"","new_string":"a"}`))
		require.Error(t, err)
		require.Contains(t, err.Error(), "old_string")
	})

	t.Run("empty new_string is allowed (deletion)", func(t *testing.T) {
		err := validateProposeEditInput(json.RawMessage(`{"description":"delete x","old_string":"x = 1\n","new_string":""}`))
		require.NoError(t, err)
	})
}

func TestValidateToolUse(t *testing.T) {
	t.Run("unregistered tool name", func(t *testing.T) {
		err := validateToolUse(&assistant.ToolUseContent{Name: "bogus", Input: json.RawMessage(`{}`)})
		require.Error(t, err)
	})

	t.Run("registered tool with invalid input", func(t *testing.T) {
		err := validateToolUse(&assistant.ToolUseContent{Name: "propose_edit", Input: json.RawMessage(`{}`)})
		require.Error(t, err)
	})

	t.Run("registered tool with valid input", func(t *testing.T) {
		err := validateToolUse(&assistant.ToolUseContent{
			Name:  "propose_edit",
			Input: json.RawMessage(`{"description":"x","old_string":"a","new_string":"b"}`),
		})
		require.NoError(t, err)
	})
}
