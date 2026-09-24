---
name: add-model
description: Add a new AI model to a provider (anthropic, openai, groq, cerebras). Use when a new model is announced and needs to be supported in aico.
argument-hint: "[provider:model-id]"
---

# Add Model Skill

Add a new AI model to the aico codebase.

## Arguments

`$ARGUMENTS` should be in the format `provider:model-id` (e.g. `anthropic:claude-opus-4-6`).
If only a model ID is given, infer the provider from the model name prefix:
- `claude-*` → `anthropic`
- `gpt-*`, `o1*`, `o3*`, `o4*` → `openai`
- `llama*`, `mixtral*` → `groq`
- `qwen*` → `cerebras`
- Other → ask the user

## Prerequisite: Gather Model Information

Before writing any code, gather the following from the user or from provided reference URLs:

1. **Model ID** — the exact string used in API calls (e.g. `claude-opus-4-6`, `gpt-4.1`)
2. **Description** — what the model is best at
3. **Pricing** — input/output per MTok
4. **Context window** and **max output tokens**
5. **Notable capabilities** (e.g. extended thinking, vision, tool calling)
6. **Predecessor model to deprecate** (optional)

If the user provides a reference URL, fetch it to extract this information.
Otherwise, search the web for the latest model information from the provider's official documentation.

## Step-by-step Procedure

### 1. Read the existing model files in the target provider

Before creating any file, **read at least one existing model file** in the same provider directory to understand the exact current pattern. The patterns described below are guidelines — always match the actual code.

Provider directories:
- `internal/providers/anthropic/`
- `internal/providers/openai/`
- `internal/providers/groq/`
- `internal/providers/cerebras/`

### 2. Create the model Go file

Create a new file at `internal/providers/<provider>/<model_file_name>.go`.

**File naming convention** — replace hyphens and dots with underscores, drop suffixes like `-versatile` or `-instant` for brevity:
- `claude-opus-4-6` → `claude_opus_4_6.go`
- `gpt-4.1` → `gpt4_1.go`
- `gpt-4.1-mini` → `gpt4_1_mini.go`
- `o3` → `o3.go`
- `llama-3.3-70b-versatile` → `llama3_3_70b.go`

**All models must:**
- Add the interface compliance checks: `var _ assistant.GenerativeModel = (*TypeName)(nil)` and `var _ assistant.GenerationOptionCapable = (*TypeName)(nil)`
- Implement all methods of `assistant.GenerativeModel`:
  - `Provider() string`
  - `Name() string`
  - `Description() string`
  - `SetSystemInstruction(...*assistant.TextContent)`
  - `GenerateContent(ctx, ...Message) (*GenerateContentResponse, error)`
  - `GenerateContentStream(ctx, ...Message) (iter.Seq2[*GenerateContentResponse, error], error)`
- Implement `assistant.GenerationOptionCapable` (`SetGenerationOptions(assistant.GenerationOptions)`), storing the value in a `genOpts` field and passing it to the request builder. This is what `[models."<name>"]` in the config file feeds; a model without it fails at runtime when settings are configured for it.

**Provider-specific patterns:**

#### Anthropic (`internal/providers/anthropic/`)

```
- Import: anthropic "github.com/anthropics/anthropic-sdk-go"
- Define: const ModelNameXxx = "model-id"
- Struct fields: systemInstruction, tools []assistant.ToolDefinition, client *anthropic.Client, genOpts assistant.GenerationOptions
- Compliance checks: GenerativeModel, ToolCapable, GenerationOptionCapable
- Constructor: takes *anthropic.Client (NOT apiKey)
- SetTools(tools ...assistant.ToolDefinition) stores m.tools
- GenerateContent: delegates to generateContent(ctx, m.client, m.Name(), m.systemInstruction, m.tools, m.genOpts, msgs)
- GenerateContentStream: delegates to generateContentStream(...) with the same arguments
```

Template: `claude_opus_5_5.go`

#### OpenAI (`internal/providers/openai/`)

```
- Struct fields: systemInstruction []*assistant.TextContent, client *APIClient, genOpts assistant.GenerationOptions
- Constructor: takes apiKey string → NewAPIClient(apiKey)
- GenerateContent: BuildChatRequest(ctx, m.Name(), m.systemInstruction, m.genOpts, msgs) → m.client.DoPost(ctx, endpoint, req, resp)
- GenerateContentStream: BuildChatRequest(...) → m.client.DoStream(ctx, endpoint, req)
- Also implements SetHttpClient(c *http.Client) (not part of interface, but required for this provider)
```

Template: `gpt4_1.go`

#### Groq / Cerebras (OpenAI-compatible)

```
- Import: "micheam.com/aico/internal/providers/openai"
- Struct fields: systemInstruction []*assistant.TextContent, client *openai.APIClient, genOpts assistant.GenerationOptions
- Constructor: takes apiKey string → openai.NewAPIClient(apiKey)
- GenerateContent: delegates to openai.GenerateContent(ctx, m.client, Endpoint, m.Name(), m.systemInstruction, m.genOpts, msgs)
- GenerateContentStream: delegates to openai.GenerateContentStream(...) with the same arguments
- Uses the provider's Endpoint constant (defined in the provider main file)
```

Template: `llama3_3_70b.go` (groq) or `gpt_oss_120b.go` (cerebras)

### 3. Register the model in the provider file

Edit the provider's main file (e.g. `anthropic.go`, `chat.go` for openai, `groq.go`, `cerebras.go`).
Update **all three** registration points:

1. **`AvailableModels()`** — add the new model struct to the returned slice
2. **`selectModel()`** — add a `case` for the model ID string
3. **`NewGenerativeModel()`** — add a `case` for the model ID with the constructor call

Place the new model at the top of each list/switch (latest model first).

If the new model is the latest of its family, also update the provider's `aliases` map (e.g. `"fable"` → the new model ID) so the family alias resolves to it.

> **Note:** `cmd/aico/models.go` does NOT need changes. It aggregates models from all providers automatically via each provider's `AvailableModels()` function.

### 4. Update the doc comment

Update the documentation comment block at the top of the provider file to include the new model's description, capabilities, and pricing.

### 5. Deprecate predecessor (if requested)

If the user requests deprecating an older model:

1. Prefix its `Description()` return with `[Deprecated] <Model Name> - superseded by <New Model>.`
2. Move its entry in the doc comment to the bottom with `(Deprecated)` suffix
3. If an alias in the provider's `aliases` map still points at it, move the alias to the successor (the alias test fails for a deprecated target)
4. Keep all code functional — do NOT remove the model

### 6. Build and test

```
make test
```

This runs `go vet ./...` and `go test -tags e2e ./...`.

### 7. Verify with CLI

Run and show the user the output of:
```
go run ./cmd/aico models
go run ./cmd/aico models describe <new-model-id>
```

If a predecessor was deprecated, also show:
```
go run ./cmd/aico models describe <deprecated-model-id>
```

### 8. Commit

Follow the Conventional Commits format:
```
feat(api): add <Model Name> support
```

If a predecessor was deprecated:
```
feat(api): add <Model Name> support and deprecate <Old Model>
```
