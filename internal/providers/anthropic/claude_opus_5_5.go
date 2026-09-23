package anthropic

import (
	"context"
	"iter"

	anthropic "github.com/anthropics/anthropic-sdk-go"

	"micheam.com/aico/internal/assistant"
)

const ModelNameClaudeOpus5_5 = "claude-opus-5-5"

type ClaudeOpus5_5 struct {
	systemInstruction []*assistant.TextContent
	tools             []assistant.ToolDefinition
	client            *anthropic.Client

	genOpts assistant.GenerationOptions
}

var (
	_ assistant.GenerativeModel         = (*ClaudeOpus5_5)(nil)
	_ assistant.ToolCapable             = (*ClaudeOpus5_5)(nil)
	_ assistant.GenerationOptionCapable = (*ClaudeOpus5_5)(nil)
)

func NewClaudeOpus5_5(client *anthropic.Client) *ClaudeOpus5_5 { return &ClaudeOpus5_5{client: client} }
func (m *ClaudeOpus5_5) Provider() string                      { return ProviderName }
func (m *ClaudeOpus5_5) Name() string                          { return ModelNameClaudeOpus5_5 }
func (m *ClaudeOpus5_5) Description() string {
	return `Claude Opus 5.5 is the latest Opus model for long-running agentic coding and
knowledge work, the successor to Claude Opus 5. Supports adaptive thinking
(always on) and effort control.
Pricing: $4/MTok input, $20/MTok output.
Supports 1M context window and 128K max output.`
}

func (m *ClaudeOpus5_5) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *ClaudeOpus5_5) SetTools(tools ...assistant.ToolDefinition) {
	m.tools = tools
}

func (m *ClaudeOpus5_5) SetGenerationOptions(opts assistant.GenerationOptions) {
	m.genOpts = opts
}

func (m *ClaudeOpus5_5) GenerateContent(
	ctx context.Context,
	msgs ...assistant.Message,
) (*assistant.GenerateContentResponse, error) {
	return generateContent(ctx, m.client, m.Name(), m.systemInstruction, m.tools, m.genOpts, msgs)
}

func (m *ClaudeOpus5_5) GenerateContentStream(
	ctx context.Context,
	msgs ...assistant.Message,
) (iter.Seq2[*assistant.GenerateContentResponse, error], error) {
	return generateContentStream(ctx, m.client, m.Name(), m.systemInstruction, m.tools, m.genOpts, msgs)
}
