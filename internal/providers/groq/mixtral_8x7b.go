package groq

import (
	"context"
	"iter"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/providers/openai"
)

type Mixtral8x7B struct {
	systemInstruction []*assistant.TextContent
	client            *openai.APIClient
}

var _ assistant.GenerativeModel = (*Mixtral8x7B)(nil)

func NewMixtral8x7B(apiKey string) *Mixtral8x7B {
	return &Mixtral8x7B{
		client: openai.NewAPIClient(apiKey),
	}
}

func (m *Mixtral8x7B) Provider() string {
	return ProviderName
}

func (m *Mixtral8x7B) Name() string {
	return "mixtral-8x7b-32768"
}

func (m *Mixtral8x7B) Description() string {
	return `Mixtral 8x7B is a Mixture of Experts model from Mistral AI with excellent performance across diverse tasks.
Powered by Groq's LPU Inference Engine for ultra-fast inference speeds.
Context window: 32K tokens. Best for: general-purpose tasks, coding, and reasoning.
Reference: https://console.groq.com/docs/models`
}

func (m *Mixtral8x7B) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *Mixtral8x7B) GenerateContent(ctx context.Context, msgs ...*assistant.Message) (*assistant.GenerateContentResponse, error) {
	return generateContent(ctx, m.client, m.Name(), m.systemInstruction, msgs)
}

func (m *Mixtral8x7B) GenerateContentStream(ctx context.Context, msgs ...*assistant.Message) (iter.Seq[*assistant.GenerateContentResponse], error) {
	return generateContentStream(ctx, m.client, m.Name(), m.systemInstruction, msgs)
}
