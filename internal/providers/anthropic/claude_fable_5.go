package anthropic

import (
	"context"
	"iter"

	anthropic "github.com/anthropics/anthropic-sdk-go"

	"micheam.com/aico/internal/assistant"
)

const ModelNameClaudeFable5 = "claude-fable-5"

type ClaudeFable5 struct {
	systemInstruction []*assistant.TextContent
	tools             []assistant.ToolDefinition
	client            *anthropic.Client

	genOpts assistant.GenerationOptions
}

var (
	_ assistant.GenerativeModel         = (*ClaudeFable5)(nil)
	_ assistant.ToolCapable             = (*ClaudeFable5)(nil)
	_ assistant.GenerationOptionCapable = (*ClaudeFable5)(nil)
)

func NewClaudeFable5(client *anthropic.Client) *ClaudeFable5 { return &ClaudeFable5{client: client} }
func (m *ClaudeFable5) Provider() string                     { return ProviderName }
func (m *ClaudeFable5) Name() string                         { return ModelNameClaudeFable5 }
func (m *ClaudeFable5) Description() string {
	return `Claude Fable 5 is Anthropic's most powerful model, excelling at
creative, agentic, and coding tasks. Supports adaptive thinking and effort control.
Pricing: $10/MTok input, $50/MTok output.
Supports 1M context window and 128K max output.`
}

func (m *ClaudeFable5) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *ClaudeFable5) SetTools(tools ...assistant.ToolDefinition) {
	m.tools = tools
}

func (m *ClaudeFable5) SetGenerationOptions(opts assistant.GenerationOptions) {
	m.genOpts = opts
}

func (m *ClaudeFable5) GenerateContent(
	ctx context.Context,
	msgs ...assistant.Message,
) (*assistant.GenerateContentResponse, error) {
	return generateContent(ctx, m.client, m.Name(), m.systemInstruction, m.tools, m.genOpts, msgs)
}

func (m *ClaudeFable5) GenerateContentStream(
	ctx context.Context,
	msgs ...assistant.Message,
) (iter.Seq2[*assistant.GenerateContentResponse, error], error) {
	return generateContentStream(ctx, m.client, m.Name(), m.systemInstruction, m.tools, m.genOpts, msgs)
}
