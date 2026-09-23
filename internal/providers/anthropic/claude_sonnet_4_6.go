package anthropic

import (
	"context"
	"iter"

	anthropic "github.com/anthropics/anthropic-sdk-go"

	"micheam.com/aico/internal/assistant"
)

const ModelNameClaudeSonnet4_6 = "claude-sonnet-4-6"

type ClaudeSonnet4_6 struct {
	systemInstruction []*assistant.TextContent
	tools             []assistant.ToolDefinition
	client            *anthropic.Client

	genOpts assistant.GenerationOptions
}

var (
	_ assistant.GenerativeModel         = (*ClaudeSonnet4_6)(nil)
	_ assistant.ToolCapable             = (*ClaudeSonnet4_6)(nil)
	_ assistant.GenerationOptionCapable = (*ClaudeSonnet4_6)(nil)
)

func NewClaudeSonnet4_6(client *anthropic.Client) *ClaudeSonnet4_6 {
	return &ClaudeSonnet4_6{client: client}
}
func (m *ClaudeSonnet4_6) Provider() string { return ProviderName }
func (m *ClaudeSonnet4_6) Name() string     { return ModelNameClaudeSonnet4_6 }
func (m *ClaudeSonnet4_6) Description() string {
	return `Claude Sonnet 4.6 is the best combination of speed and intelligence.
Supports extended thinking and adaptive thinking. Fast comparative latency.
Pricing: $3/MTok input, $15/MTok output.
Supports 200K context window (1M with beta header) and 64K max output.`
}

func (m *ClaudeSonnet4_6) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *ClaudeSonnet4_6) SetTools(tools ...assistant.ToolDefinition) {
	m.tools = tools
}

func (m *ClaudeSonnet4_6) SetGenerationOptions(opts assistant.GenerationOptions) {
	m.genOpts = opts
}

func (m *ClaudeSonnet4_6) GenerateContent(
	ctx context.Context,
	msgs ...assistant.Message,
) (*assistant.GenerateContentResponse, error) {
	return generateContent(ctx, m.client, m.Name(), m.systemInstruction, m.tools, m.genOpts, msgs)
}

func (m *ClaudeSonnet4_6) GenerateContentStream(
	ctx context.Context,
	msgs ...assistant.Message,
) (iter.Seq2[*assistant.GenerateContentResponse, error], error) {
	return generateContentStream(ctx, m.client, m.Name(), m.systemInstruction, m.tools, m.genOpts, msgs)
}
