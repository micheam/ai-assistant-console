package anthropic

import (
	"context"
	"iter"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	anthropicopt "github.com/anthropics/anthropic-sdk-go/option"

	"micheam.com/aico/internal/assistant"
)

const ModelNameClaudeSonnet5 = "claude-sonnet-5"

type ClaudeSonnet5 struct {
	systemInstruction []*assistant.TextContent
	tools             []assistant.ToolDefinition
	client            *anthropic.Client

	opts []anthropicopt.RequestOption
}

var (
	_ assistant.GenerativeModel = (*ClaudeSonnet5)(nil)
	_ assistant.ToolCapable     = (*ClaudeSonnet5)(nil)
)

func NewClaudeSonnet5(client *anthropic.Client) *ClaudeSonnet5 { return &ClaudeSonnet5{client: client} }
func (m *ClaudeSonnet5) Provider() string                      { return ProviderName }
func (m *ClaudeSonnet5) Name() string                          { return ModelNameClaudeSonnet5 }
func (m *ClaudeSonnet5) Description() string {
	return `Claude Sonnet 5 is the best combination of speed and intelligence,
the successor to Claude Sonnet 4.6. Supports adaptive thinking and effort control.
Pricing: $3/MTok input, $15/MTok output.
Supports 1M context window and 128K max output.`
}

func (m *ClaudeSonnet5) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *ClaudeSonnet5) SetTools(tools ...assistant.ToolDefinition) {
	m.tools = tools
}

func (m *ClaudeSonnet5) GenerateContent(
	ctx context.Context,
	msgs ...assistant.Message,
) (*assistant.GenerateContentResponse, error) {
	return generateContent(ctx, m.client, m.Name(), m.systemInstruction, m.tools, m.opts, msgs)
}

func (m *ClaudeSonnet5) GenerateContentStream(
	ctx context.Context,
	msgs ...assistant.Message,
) (iter.Seq2[*assistant.GenerateContentResponse, error], error) {
	return generateContentStream(ctx, m.client, m.Name(), m.systemInstruction, m.tools, m.opts, msgs)
}
