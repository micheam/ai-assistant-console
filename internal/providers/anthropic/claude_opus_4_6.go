package anthropic

import (
	"context"
	"iter"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	anthropicopt "github.com/anthropics/anthropic-sdk-go/option"

	"micheam.com/aico/internal/assistant"
)

const ModelNameClaudeOpus4_6 = "claude-opus-4-6"

type ClaudeOpus4_6 struct {
	systemInstruction []*assistant.TextContent
	tools             []assistant.ToolDefinition
	client            *anthropic.Client

	opts []anthropicopt.RequestOption
}

var (
	_ assistant.GenerativeModel = (*ClaudeOpus4_6)(nil)
	_ assistant.ToolCapable     = (*ClaudeOpus4_6)(nil)
)

func NewClaudeOpus4_6(client *anthropic.Client) *ClaudeOpus4_6 { return &ClaudeOpus4_6{client: client} }
func (m *ClaudeOpus4_6) Provider() string                      { return ProviderName }
func (m *ClaudeOpus4_6) Name() string                          { return ModelNameClaudeOpus4_6 }
func (m *ClaudeOpus4_6) Description() string {
	return `[Deprecated] Claude Opus 4.6 - superseded by Claude Opus 4.8.
Claude Opus 4.6 is the most intelligent model for building agents and coding.
Top-tier results in reasoning, coding, multilingual tasks, and long-context handling.
Supports extended thinking and adaptive thinking. Pricing: $5/MTok input, $25/MTok output.
Supports 200K context window (1M with beta header) and 128K max output.`
}

func (m *ClaudeOpus4_6) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *ClaudeOpus4_6) SetTools(tools ...assistant.ToolDefinition) {
	m.tools = tools
}

func (m *ClaudeOpus4_6) GenerateContent(
	ctx context.Context,
	msgs ...assistant.Message,
) (*assistant.GenerateContentResponse, error) {
	return generateContent(ctx, m.client, m.Name(), m.systemInstruction, m.tools, m.opts, msgs)
}

func (m *ClaudeOpus4_6) GenerateContentStream(
	ctx context.Context,
	msgs ...assistant.Message,
) (iter.Seq2[*assistant.GenerateContentResponse, error], error) {
	return generateContentStream(ctx, m.client, m.Name(), m.systemInstruction, m.tools, m.opts, msgs)
}
