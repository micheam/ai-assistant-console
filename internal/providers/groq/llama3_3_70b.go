package groq

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
	return "llama-3.3-70b-versatile"
}

func (m *Llama3_3_70B) Description() string {
	return `Llama 3.3 70B Versatile is Meta's latest large language model optimized for versatile tasks.
Powered by Groq's LPU Inference Engine for ultra-fast inference speeds (~500-800 tokens/sec).
Context window: 128K tokens. Best for: complex reasoning, coding, and creative tasks.
Reference: https://console.groq.com/docs/models`
}

func (m *Llama3_3_70B) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *Llama3_3_70B) GenerateContent(ctx context.Context, msgs ...*assistant.Message) (*assistant.GenerateContentResponse, error) {
	return openai.GenerateContent(ctx, m.client, Endpoint, m.Name(), m.systemInstruction, msgs)
}

func (m *Llama3_3_70B) GenerateContentStream(ctx context.Context, msgs ...*assistant.Message) (iter.Seq[*assistant.GenerateContentResponse], error) {
	return openai.GenerateContentStream(ctx, m.client, Endpoint, m.Name(), m.systemInstruction, msgs)
}
