package cerebras

import (
	"context"
	"iter"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/providers/openai"
)

type Llama3_1_8B struct {
	systemInstruction []*assistant.TextContent
	client            *openai.APIClient
}

var _ assistant.GenerativeModel = (*Llama3_1_8B)(nil)

func NewLlama3_1_8B(apiKey string) *Llama3_1_8B {
	return &Llama3_1_8B{
		client: openai.NewAPIClient(apiKey),
	}
}

func (m *Llama3_1_8B) Provider() string {
	return ProviderName
}

func (m *Llama3_1_8B) Name() string {
	return "llama-3.1-8b"
}

func (m *Llama3_1_8B) Description() string {
	return `Llama 3.1 8B on Cerebras Inference delivers ultra-fast inference (~1800-2200 tokens/sec).
A lightweight model ideal for quick tasks while maintaining Cerebras' signature speed advantage.
Best for: simple tasks, quick Q&A, and applications requiring minimal latency.
Reference: https://inference-docs.cerebras.ai/`
}

func (m *Llama3_1_8B) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *Llama3_1_8B) GenerateContent(ctx context.Context, msgs ...*assistant.Message) (*assistant.GenerateContentResponse, error) {
	return generateContent(ctx, m.client, m.Name(), m.systemInstruction, msgs)
}

func (m *Llama3_1_8B) GenerateContentStream(ctx context.Context, msgs ...*assistant.Message) (iter.Seq[*assistant.GenerateContentResponse], error) {
	return generateContentStream(ctx, m.client, m.Name(), m.systemInstruction, msgs)
}
