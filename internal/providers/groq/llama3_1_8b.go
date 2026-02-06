package groq

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
	return "llama-3.1-8b-instant"
}

func (m *Llama3_1_8B) Description() string {
	return `Llama 3.1 8B Instant is a lightweight, fast model from Meta optimized for quick responses.
Powered by Groq's LPU Inference Engine for ultra-fast inference speeds.
Context window: 128K tokens. Best for: simple tasks, quick Q&A, and low-latency applications.
Reference: https://console.groq.com/docs/models`
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
