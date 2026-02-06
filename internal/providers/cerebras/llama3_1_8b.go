package cerebras

import (
	"context"
	"iter"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/providers/openai"
)

type Llama3_1_8B struct {
	assistant.DeprecationInfo
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
	return "llama3.1-8b"
}

func (m *Llama3_1_8B) Description() string {
	return `Llama 3.1 8B - Lightweight and fast (~2200 tokens/sec).
Best for: Simple question answering and lightweight tasks.
Reference: https://inference-docs.cerebras.ai/models/llama-31-8b.md`
}

func (m *Llama3_1_8B) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *Llama3_1_8B) GenerateContent(ctx context.Context, msgs ...*assistant.Message) (*assistant.GenerateContentResponse, error) {
	return openai.GenerateContent(ctx, m.client, Endpoint, m.Name(), m.systemInstruction, msgs)
}

func (m *Llama3_1_8B) GenerateContentStream(ctx context.Context, msgs ...*assistant.Message) (iter.Seq[*assistant.GenerateContentResponse], error) {
	return openai.GenerateContentStream(ctx, m.client, Endpoint, m.Name(), m.systemInstruction, msgs)
}
