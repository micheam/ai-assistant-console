package anthropic

import (
	"context"
	"iter"

	anthropic "github.com/anthropics/anthropic-sdk-go"

	"micheam.com/aico/internal/assistant"
)

const ModelNameClaudeOpus4_8 = "claude-opus-4-8"

type ClaudeOpus4_8 struct {
	systemInstruction []*assistant.TextContent
	tools             []assistant.ToolDefinition
	client            *anthropic.Client

	genOpts assistant.GenerationOptions
}

var (
	_ assistant.GenerativeModel         = (*ClaudeOpus4_8)(nil)
	_ assistant.ToolCapable             = (*ClaudeOpus4_8)(nil)
	_ assistant.GenerationOptionCapable = (*ClaudeOpus4_8)(nil)
)

func NewClaudeOpus4_8(client *anthropic.Client) *ClaudeOpus4_8 { return &ClaudeOpus4_8{client: client} }
func (m *ClaudeOpus4_8) Provider() string                      { return ProviderName }
func (m *ClaudeOpus4_8) Name() string                          { return ModelNameClaudeOpus4_8 }
func (m *ClaudeOpus4_8) Description() string {
	return `Claude Opus 4.8 is the latest Opus model for building agents and coding.
Top-tier results in reasoning, coding, multilingual tasks, and long-context handling.
Supports adaptive thinking and effort control. Pricing: $5/MTok input, $25/MTok output.
Supports 1M context window and 128K max output.`
}

func (m *ClaudeOpus4_8) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *ClaudeOpus4_8) SetTools(tools ...assistant.ToolDefinition) {
	m.tools = tools
}

func (m *ClaudeOpus4_8) SetGenerationOptions(opts assistant.GenerationOptions) {
	m.genOpts = opts
}

func (m *ClaudeOpus4_8) GenerateContent(
	ctx context.Context,
	msgs ...assistant.Message,
) (*assistant.GenerateContentResponse, error) {
	return generateContent(ctx, m.client, m.Name(), m.systemInstruction, m.tools, m.genOpts, msgs)
}

func (m *ClaudeOpus4_8) GenerateContentStream(
	ctx context.Context,
	msgs ...assistant.Message,
) (iter.Seq2[*assistant.GenerateContentResponse, error], error) {
	return generateContentStream(ctx, m.client, m.Name(), m.systemInstruction, m.tools, m.genOpts, msgs)
}
