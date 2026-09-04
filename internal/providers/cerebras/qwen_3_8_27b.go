package cerebras

import (
	"context"
	"iter"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/providers/openai"
)

type Qwen3_8_27B struct {
	systemInstruction []*assistant.TextContent
	client            *openai.APIClient
}

var _ assistant.GenerativeModel = (*Qwen3_8_27B)(nil)

func NewQwen3_8_27B(apiKey string) *Qwen3_8_27B {
	return &Qwen3_8_27B{
		client: openai.NewAPIClient(apiKey),
	}
}

func (m *Qwen3_8_27B) Provider() string {
	return ProviderName
}

func (m *Qwen3_8_27B) Name() string {
	return "qwen-3.8-27b"
}

func (m *Qwen3_8_27B) Description() string {
	return `Qwen 3.8 27B - Alibaba's 27B dense multimodal model for agentic coding, tool use, research, and long-running workflows.
Best for: Complex reasoning and multimodal tasks with extended thinking enabled by default.
Pricing: $0.99/MTok input, $1.49/MTok output. Context: 128K tokens (paid), 64K tokens (free trial).
Reference: https://inference-docs.cerebras.ai/models/qwen-3.8-27b`
}

func (m *Qwen3_8_27B) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *Qwen3_8_27B) GenerateContent(ctx context.Context, msgs ...assistant.Message) (*assistant.GenerateContentResponse, error) {
	return openai.GenerateContent(ctx, m.client, Endpoint, m.Name(), m.systemInstruction, msgs)
}

func (m *Qwen3_8_27B) GenerateContentStream(ctx context.Context, msgs ...assistant.Message) (iter.Seq2[*assistant.GenerateContentResponse, error], error) {
	return openai.GenerateContentStream(ctx, m.client, Endpoint, m.Name(), m.systemInstruction, msgs)
}
