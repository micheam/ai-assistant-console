package anthropic

import (
	"context"
	"iter"

	anthropic "github.com/anthropics/anthropic-sdk-go"

	"micheam.com/aico/internal/assistant"
)

const ModelNameClaudeOpus5 = "claude-opus-5"

type ClaudeOpus5 struct {
	systemInstruction []*assistant.TextContent
	tools             []assistant.ToolDefinition
	client            *anthropic.Client

	genOpts assistant.GenerationOptions
}

var (
	_ assistant.GenerativeModel         = (*ClaudeOpus5)(nil)
	_ assistant.ToolCapable             = (*ClaudeOpus5)(nil)
	_ assistant.GenerationOptionCapable = (*ClaudeOpus5)(nil)
)

func NewClaudeOpus5(client *anthropic.Client) *ClaudeOpus5 { return &ClaudeOpus5{client: client} }
func (m *ClaudeOpus5) Provider() string                    { return ProviderName }
func (m *ClaudeOpus5) Name() string                        { return ModelNameClaudeOpus5 }
func (m *ClaudeOpus5) Description() string {
	return `Claude Opus 5 is the Opus model for long-running agentic coding and
knowledge work, the successor to Claude Opus 4.8. Supports adaptive thinking
(on by default) and effort control.
Pricing: $5/MTok input, $25/MTok output.
Supports 1M context window and 128K max output.`
}

func (m *ClaudeOpus5) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *ClaudeOpus5) SetTools(tools ...assistant.ToolDefinition) {
	m.tools = tools
}

func (m *ClaudeOpus5) SetGenerationOptions(opts assistant.GenerationOptions) {
	m.genOpts = opts
}

func (m *ClaudeOpus5) GenerateContent(
	ctx context.Context,
	msgs ...assistant.Message,
) (*assistant.GenerateContentResponse, error) {
	return generateContent(ctx, m.client, m.Name(), m.systemInstruction, m.tools, m.genOpts, msgs)
}

func (m *ClaudeOpus5) GenerateContentStream(
	ctx context.Context,
	msgs ...assistant.Message,
) (iter.Seq2[*assistant.GenerateContentResponse, error], error) {
	return generateContentStream(ctx, m.client, m.Name(), m.systemInstruction, m.tools, m.genOpts, msgs)
}
