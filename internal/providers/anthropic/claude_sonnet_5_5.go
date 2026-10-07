package anthropic

import (
	"context"
	"iter"

	anthropic "github.com/anthropics/anthropic-sdk-go"

	"micheam.com/aico/internal/assistant"
)

const ModelNameClaudeSonnet5_5 = "claude-sonnet-5-5"

type ClaudeSonnet5_5 struct {
	systemInstruction []*assistant.TextContent
	tools             []assistant.ToolDefinition
	client            *anthropic.Client

	genOpts assistant.GenerationOptions
}

var (
	_ assistant.GenerativeModel         = (*ClaudeSonnet5_5)(nil)
	_ assistant.ToolCapable             = (*ClaudeSonnet5_5)(nil)
	_ assistant.GenerationOptionCapable = (*ClaudeSonnet5_5)(nil)
)

func NewClaudeSonnet5_5(client *anthropic.Client) *ClaudeSonnet5_5 {
	return &ClaudeSonnet5_5{client: client}
}

func (m *ClaudeSonnet5_5) Provider() string { return ProviderName }
func (m *ClaudeSonnet5_5) Name() string     { return ModelNameClaudeSonnet5_5 }
func (m *ClaudeSonnet5_5) Description() string {
	return `Claude Sonnet 5.5 is the best combination of speed and intelligence,
the successor to Claude Sonnet 5. Supports adaptive thinking and effort control.
Pricing: $2/MTok input, $10/MTok output.
Supports 1M context window and 128K max output.`
}

func (m *ClaudeSonnet5_5) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *ClaudeSonnet5_5) SetTools(tools ...assistant.ToolDefinition) {
	m.tools = tools
}

func (m *ClaudeSonnet5_5) SetGenerationOptions(opts assistant.GenerationOptions) {
	m.genOpts = opts
}

func (m *ClaudeSonnet5_5) GenerateContent(
	ctx context.Context,
	msgs ...assistant.Message,
) (*assistant.GenerateContentResponse, error) {
	return generateContent(ctx, m.client, m.Name(), m.systemInstruction, m.tools, m.genOpts, msgs)
}

func (m *ClaudeSonnet5_5) GenerateContentStream(
	ctx context.Context,
	msgs ...assistant.Message,
) (iter.Seq2[*assistant.GenerateContentResponse, error], error) {
	return generateContentStream(ctx, m.client, m.Name(), m.systemInstruction, m.tools, m.genOpts, msgs)
}
