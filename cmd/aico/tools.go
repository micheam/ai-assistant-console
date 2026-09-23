package main

import (
	"encoding/json"
	"fmt"
	"slices"

	"micheam.com/aico/internal/assistant"
)

const toolNameProposeEdit = "propose_edit"

// proposeEditDescription tells the model what propose_edit does and, more
// importantly, when to call it. This text (and proposeEditInputSchema) is
// part of the byte-identical request prefix that preserved-thinking
// accounts require to stay unchanged across a session's later turns:
// changing it is a breaking change for sessions that already used the
// tool. Version the tool name (e.g. propose_edit_v2) instead of editing
// this text in place, or accept that older sessions can only resume with
// thinking blocks dropped.
const proposeEditDescription = `Propose a concrete edit to the text inside the user's <source> block. The edit is NOT
applied automatically: it is shown to the user, who decides whether to apply it in
their editor.

Call this tool whenever your answer includes a specific change to the <source> text
(a fix, a rename, a refactor, a rewritten sentence, an added or removed block). Do not
describe such a change only in prose or in a diff code block; express it with this
tool instead. Your explanation of WHY goes in normal text written BEFORE the tool
call(s), because your turn ends when you call this tool.

Input rules:
- old_string: the exact text to replace, copied verbatim from <source> (same
  whitespace, indentation and line breaks). It must occur exactly once in <source>;
  include surrounding lines if needed to make it unique. Must not be empty.
- new_string: the replacement text. May be empty to delete old_string.
- description: one short line summarizing the change.

Keep each proposal small and self-contained (one logical change). For several
changes, call this tool once per change. Proposals must not overlap each other,
because the user may apply only some of them, in any order. If no concrete change
to <source> is warranted, do not call this tool.`

var proposeEditTool = assistant.ToolDefinition{
	Name:        toolNameProposeEdit,
	Description: proposeEditDescription,
	InputSchema: map[string]any{
		"type": "object",
		"properties": map[string]any{
			"description": map[string]any{
				"type":        "string",
				"description": "One short line summarizing the change.",
			},
			"old_string": map[string]any{
				"type":        "string",
				"description": "The exact text to replace, copied verbatim from <source>. Must occur exactly once.",
			},
			"new_string": map[string]any{
				"type":        "string",
				"description": "The replacement text. May be empty to delete old_string.",
			},
		},
		"required": []string{"description", "old_string", "new_string"},
	},
}

// toolRegistry maps an enableable tool's name to its definition and input
// validator. It is the single source of truth for which --tool values are
// accepted.
var toolRegistry = map[string]struct {
	def      assistant.ToolDefinition
	validate func(json.RawMessage) error
}{
	toolNameProposeEdit: {def: proposeEditTool, validate: validateProposeEditInput},
}

// resolveTools turns a set of tool names into their definitions, in a
// stable (sorted) order so the request's tools array is deterministic
// across turns. It errors on any name not in toolRegistry.
func resolveTools(names []string) ([]assistant.ToolDefinition, error) {
	seen := make(map[string]bool, len(names))
	var unique []string
	for _, n := range names {
		if n == "" || seen[n] {
			continue
		}
		seen[n] = true
		unique = append(unique, n)
	}
	slices.Sort(unique)

	defs := make([]assistant.ToolDefinition, 0, len(unique))
	for _, n := range unique {
		entry, ok := toolRegistry[n]
		if !ok {
			return nil, fmt.Errorf("unknown tool %q (available: propose_edit)", n)
		}
		defs = append(defs, entry.def)
	}
	return defs, nil
}

// unionToolNames merges two lists of tool names into a sorted, deduplicated
// list, used to combine a session's already-persisted tools with any new
// ones passed via --tool on this invocation.
func unionToolNames(a, b []string) []string {
	seen := make(map[string]bool, len(a)+len(b))
	var out []string
	for _, n := range append(append([]string{}, a...), b...) {
		if n == "" || seen[n] {
			continue
		}
		seen[n] = true
		out = append(out, n)
	}
	slices.Sort(out)
	return out
}

// validateToolUse checks a tool_use content block returned by the model
// against its registered tool's input schema, before it is shown to the
// user as a proposal or persisted to the session.
func validateToolUse(tu *assistant.ToolUseContent) error {
	entry, ok := toolRegistry[tu.Name]
	if !ok {
		return fmt.Errorf("model called unregistered tool %q", tu.Name)
	}
	return entry.validate(tu.Input)
}

// proposeEditFields extracts description/old_string/new_string from a
// propose_edit tool_use input, for rendering. ok is false if input doesn't
// unmarshal into the expected shape (the caller should fall back to
// printing the raw input).
func proposeEditFields(input json.RawMessage) (description, oldString, newString string, ok bool) {
	var fields struct {
		Description string `json:"description"`
		OldString   string `json:"old_string"`
		NewString   string `json:"new_string"`
	}
	if err := json.Unmarshal(input, &fields); err != nil {
		return "", "", "", false
	}
	return fields.Description, fields.OldString, fields.NewString, true
}

// validateProposeEditInput validates a propose_edit tool_use input: valid
// JSON, all three fields present as strings, and a non-empty old_string.
func validateProposeEditInput(input json.RawMessage) error {
	if !json.Valid(input) {
		return fmt.Errorf("propose_edit input is not valid JSON")
	}
	var fields struct {
		Description *string `json:"description"`
		OldString   *string `json:"old_string"`
		NewString   *string `json:"new_string"`
	}
	if err := json.Unmarshal(input, &fields); err != nil {
		return fmt.Errorf("propose_edit input: %w", err)
	}
	if fields.Description == nil {
		return fmt.Errorf("propose_edit input missing %q", "description")
	}
	if fields.OldString == nil {
		return fmt.Errorf("propose_edit input missing %q", "old_string")
	}
	if fields.NewString == nil {
		return fmt.Errorf("propose_edit input missing %q", "new_string")
	}
	if *fields.OldString == "" {
		return fmt.Errorf("propose_edit input: %q must not be empty", "old_string")
	}
	return nil
}
