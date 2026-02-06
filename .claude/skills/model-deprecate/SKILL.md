---
name: model-deprecate
description: Mark an existing model as deprecated. Sets structured deprecation metadata and removed-in version.
argument-hint: "[provider:model-id]"
---

# Model Deprecate Skill

Mark an existing AI model as deprecated in the aico codebase.

## Arguments

`$ARGUMENTS` should be in the format `provider:model-id` (e.g. `anthropic:claude-opus-4-5`).
If only a model ID is given, infer the provider from the model name prefix:
- `claude-*` → `anthropic`
- `gpt-*`, `o1*`, `o3*`, `o4*` → `openai`
- `llama*`, `mixtral*` → `groq`
- Other → ask the user

## Prerequisite: Gather Information

Before making changes, confirm the following with the user:

1. **Model ID** to deprecate (e.g. `claude-opus-4-5`)
2. **Removed-in version** — the version in which this model will be removed (e.g. `v2.0.0`)

## Step-by-step Procedure

### 1. Locate the model file

Find the model implementation file at `internal/providers/<provider>/`.
Use the model ID to find the matching Go file (e.g. `claude-opus-4-5` → `claude_opus_4_5.go`).

### 2. Set DeprecationInfo in AvailableModels()

Edit the provider's main file (`internal/providers/<provider>/<provider>.go` or equivalent).

In the `AvailableModels()` function, update the model's struct literal to include deprecation metadata:

```go
// Before:
&ClaudeOpus4_5{},

// After:
&ClaudeOpus4_5{DeprecationInfo: assistant.DeprecationInfo{IsDeprecated: true, RemovedIn: "v2.0.0"}},
```

### 3. Update Description() (if not already deprecated)

If the model's `Description()` method does not already have a `[Deprecated]` prefix, update it:

```go
func (m *ModelName) Description() string {
    return `[Deprecated] Model Name - superseded by <New Model>.
... existing description ...`
}
```

### 4. Update the doc comment

Move the model's entry in the provider file's doc comment to the bottom and add `(Deprecated)` suffix.

### 5. Build and test

Run:
```
go build ./...
go test ./...
```

### 6. Verify with CLI

Run and show the user the output of:
```
go run ./cmd/aico models describe <deprecated-model-id>
```

Confirm the output includes:
- `Deprecated: yes`
- `Removed In: vX.Y.Z`

Also verify the model is filtered from default list:
```
go run ./cmd/aico models
go run ./cmd/aico models --all
```

### 7. Commit

Follow the Conventional Commits format:
```
chore(api): deprecate <model-name>
```
