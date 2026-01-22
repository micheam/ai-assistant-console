package cerebras

import (
	"context"
	"iter"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/providers/openai"
)

type Llama3_3_70B struct {
	systemInstruction []*assistant.TextContent
	client            *openai.APIClient
}

var _ assistant.GenerativeModel = (*Llama3_3_70B)(nil)

func NewLlama3_3_70B(apiKey string) *Llama3_3_70B {
	return &Llama3_3_70B{
		client: openai.NewAPIClient(apiKey),
	}
}

func (m *Llama3_3_70B) Provider() string {
	return ProviderName
}

func (m *Llama3_3_70B) Name() string {
	return "llama-3.3-70b"
}

func (m *Llama3_3_70B) Description() string {
	return `Llama 3.3 70B on Cerebras Inference delivers blazing fast inference (~2000-2500 tokens/sec).
Cerebras' Wafer-Scale Engine eliminates memory bandwidth bottlenecks by keeping the entire model on-chip.
Best for: tasks requiring extremely fast response times, complex reasoning, and code generation.
Reference: https://inference-docs.cerebras.ai/`
}

func (m *Llama3_3_70B) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *Llama3_3_70B) GenerateContent(ctx context.Context, msgs ...*assistant.Message) (*assistant.GenerateContentResponse, error) {
	return generateContent(ctx, m.client, m.Name(), m.systemInstruction, msgs)
}

func (m *Llama3_3_70B) GenerateContentStream(ctx context.Context, msgs ...*assistant.Message) (iter.Seq[*assistant.GenerateContentResponse], error) {
	return generateContentStream(ctx, m.client, m.Name(), m.systemInstruction, msgs)
}
