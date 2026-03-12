---
name: add-provider
description: Add a new AI provider to aico (e.g. a new OpenAI-compatible API service). Scaffolds provider package and registers it in the CLI.
argument-hint: "[provider-name]"
---

# Add Provider Skill

Scaffold a new AI provider package and integrate it into the aico CLI.

## Arguments

`$ARGUMENTS` is the provider name (e.g. `together`, `fireworks`, `deepseek`).

## Prerequisite: Gather Provider Information

Before writing any code, gather the following:

1. **Provider name** — lowercase, used as package name and directory name
2. **API endpoint** — the chat completions URL
3. **API compatibility** — OpenAI-compatible or custom SDK
4. **API key environment variable** — e.g. `AICO_DEEPSEEK_API_KEY`
5. **Initial models** — at least 1 model to include (model ID, description, pricing)

If the user provides a reference URL, fetch it.
Otherwise, search the web for the provider's API documentation.

## Step-by-step Procedure

### 1. Read existing provider implementations

Read at least one complete provider to understand the pattern:
- **OpenAI-compatible**: read `internal/providers/groq/groq.go` + one model file (e.g. `llama3_3_70b.go`)
- **Custom SDK**: read `internal/providers/anthropic/anthropic.go` + one model file

Most new providers will be OpenAI-compatible (same API shape as Groq/Cerebras).

### 2. Create the provider package

Create `internal/providers/<provider>/` with:

#### `<provider>.go` — Provider main file

Must contain:
- `const Endpoint = "https://..."` — API endpoint
- `const ProviderName = "<provider>"` — provider identifier
- `func AvailableModels() []assistant.ModelDescriptor` — returns all model structs
- `func DescribeModel(modelName string) (desc string, found bool)` — lookup by name
- `func selectModel(modelName string) (assistant.GenerativeModel, bool)` — internal switch
- `func NewGenerativeModel(modelName, apiKey string) (assistant.GenerativeModel, error)` — factory

Template: `internal/providers/groq/groq.go`

#### Model files — one per model

Follow the same pattern as the `add-model` skill. For OpenAI-compatible providers:

```go
package <provider>

import (
    "context"
    "iter"

    "micheam.com/aico/internal/assistant"
    "micheam.com/aico/internal/providers/openai"
)

type ModelStruct struct {
    systemInstruction []*assistant.TextContent
    client            *openai.APIClient
}

var _ assistant.GenerativeModel = (*ModelStruct)(nil)

func NewModelStruct(apiKey string) *ModelStruct {
    return &ModelStruct{client: openai.NewAPIClient(apiKey)}
}

func (m *ModelStruct) Provider() string    { return ProviderName }
func (m *ModelStruct) Name() string        { return "model-id" }
func (m *ModelStruct) Description() string { return `...` }

func (m *ModelStruct) SetSystemInstruction(contents ...*assistant.TextContent) {
    m.systemInstruction = contents
}

func (m *ModelStruct) GenerateContent(ctx context.Context, msgs ...assistant.Message) (*assistant.GenerateContentResponse, error) {
    return openai.GenerateContent(ctx, m.client, Endpoint, m.Name(), m.systemInstruction, msgs)
}

func (m *ModelStruct) GenerateContentStream(ctx context.Context, msgs ...assistant.Message) (iter.Seq2[*assistant.GenerateContentResponse, error], error) {
    return openai.GenerateContentStream(ctx, m.client, Endpoint, m.Name(), m.systemInstruction, msgs)
}
```

### 3. Register the provider in CLI

Edit `cmd/aico/models.go`:

1. **Add import**: `"micheam.com/aico/internal/providers/<provider>"`
2. **`allAvailableModels()`**: append `<provider>.AvailableModels()...`
3. **`validateProviderModel()`**: add `case <provider>.ProviderName:`
4. **`detectProviderByModelSpec()`**: add to the providers search order slice

Edit `cmd/aico/main.go`:

1. **Add API key flag**: define `flagAPIKey<Provider>` with env source `AICO_<PROVIDER>_API_KEY`
2. **`detectModel()`**: add `case <provider>.ProviderName:` to the provider switch

### 4. Update provider prefix detection (if applicable)

In the `add-model` skill's argument parsing and in any documentation, note the new provider's model name prefix pattern.

### 5. Build and test

```
make test
```

### 6. Verify with CLI

```
go run ./cmd/aico models
go run ./cmd/aico models describe <new-model-id>
```

### 7. Commit

```
feat(api): add <Provider Name> provider with <Model Name> support
```
