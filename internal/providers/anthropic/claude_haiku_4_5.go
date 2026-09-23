package anthropic

import (
	"context"
	"iter"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	anthropicopt "github.com/anthropics/anthropic-sdk-go/option"

	"micheam.com/aico/internal/assistant"
)

const ModelNameClaudeHaiku4_5 = "claude-haiku-4-5"

type ClaudeHaiku4_5 struct {
	systemInstruction []*assistant.TextContent
	tools             []assistant.ToolDefinition
	client            *anthropic.Client

	opts []anthropicopt.RequestOption
}

var (
	_ assistant.GenerativeModel = (*ClaudeHaiku4_5)(nil)
	_ assistant.ToolCapable     = (*ClaudeHaiku4_5)(nil)
)

func NewClaudeHaiku4_5(client *anthropic.Client) *ClaudeHaiku4_5 {
	return &ClaudeHaiku4_5{client: client}
}

func (m *ClaudeHaiku4_5) Provider() string { return ProviderName }
func (m *ClaudeHaiku4_5) Name() string     { return ModelNameClaudeHaiku4_5 }
func (m *ClaudeHaiku4_5) Description() string {
	return `Claude Haiku 4.5 is the fastest model with near-frontier performance.
Engineered for lightning-fast speed at the most economical price point.
Best for real-time applications, high-volume intelligent processing,
cost-sensitive deployments needing strong reasoning, and sub-agent tasks.
Supports 200K context window.`
}

func (m *ClaudeHaiku4_5) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *ClaudeHaiku4_5) SetTools(tools ...assistant.ToolDefinition) {
	m.tools = tools
}

func (m *ClaudeHaiku4_5) GenerateContent(
	ctx context.Context,
	msgs ...assistant.Message,
) (*assistant.GenerateContentResponse, error) {
	return generateContent(ctx, m.client, m.Name(), m.systemInstruction, m.tools, m.opts, msgs)
}

func (m *ClaudeHaiku4_5) GenerateContentStream(ctx context.Context, msgs ...assistant.Message) (iter.Seq2[*assistant.GenerateContentResponse, error], error) {
	return generateContentStream(ctx, m.client, m.Name(), m.systemInstruction, m.tools, m.opts, msgs)
}
