package anthropic

import (
	"context"
	"iter"

	anthropic "github.com/anthropics/anthropic-sdk-go"

	"micheam.com/aico/internal/assistant"
)

const ModelNameClaudeFable5_1 = "claude-fable-5-1"

type ClaudeFable5_1 struct {
	systemInstruction []*assistant.TextContent
	tools             []assistant.ToolDefinition
	client            *anthropic.Client

	genOpts assistant.GenerationOptions
}

var (
	_ assistant.GenerativeModel         = (*ClaudeFable5_1)(nil)
	_ assistant.ToolCapable             = (*ClaudeFable5_1)(nil)
	_ assistant.GenerationOptionCapable = (*ClaudeFable5_1)(nil)
)

func NewClaudeFable5_1(client *anthropic.Client) *ClaudeFable5_1 {
	return &ClaudeFable5_1{client: client}
}
func (m *ClaudeFable5_1) Provider() string { return ProviderName }
func (m *ClaudeFable5_1) Name() string     { return ModelNameClaudeFable5_1 }
func (m *ClaudeFable5_1) Description() string {
	return `Claude Fable 5.1 is Anthropic's model for demanding reasoning and
long-horizon agentic work, the successor to Claude Fable 5. Supports adaptive
thinking (always on) and effort control.
Pricing: $10/MTok input, $50/MTok output.
Supports 1M context window and 128K max output.`
}

func (m *ClaudeFable5_1) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *ClaudeFable5_1) SetTools(tools ...assistant.ToolDefinition) {
	m.tools = tools
}

func (m *ClaudeFable5_1) SetGenerationOptions(opts assistant.GenerationOptions) {
	m.genOpts = opts
}

func (m *ClaudeFable5_1) GenerateContent(
	ctx context.Context,
	msgs ...assistant.Message,
) (*assistant.GenerateContentResponse, error) {
	return generateContent(ctx, m.client, m.Name(), m.systemInstruction, m.tools, m.genOpts, msgs)
}

func (m *ClaudeFable5_1) GenerateContentStream(
	ctx context.Context,
	msgs ...assistant.Message,
) (iter.Seq2[*assistant.GenerateContentResponse, error], error) {
	return generateContentStream(ctx, m.client, m.Name(), m.systemInstruction, m.tools, m.genOpts, msgs)
}
