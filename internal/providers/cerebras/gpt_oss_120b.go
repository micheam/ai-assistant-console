package cerebras

import (
	"context"
	"iter"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/providers/openai"
)

type GptOss120B struct {
	systemInstruction []*assistant.TextContent
	client            *openai.APIClient
}

var _ assistant.GenerativeModel = (*GptOss120B)(nil)

func NewGptOss120B(apiKey string) *GptOss120B {
	return &GptOss120B{
		client: openai.NewAPIClient(apiKey),
	}
}

func (m *GptOss120B) Provider() string {
	return ProviderName
}

func (m *GptOss120B) Name() string {
	return "gpt-oss-120b"
}

func (m *GptOss120B) Description() string {
	return `GPT-OSS 120B - General-purpose and fastest model (~3000 tokens/sec).
Best for: File summarization, explanations, and general code generation.
Reference: https://inference-docs.cerebras.ai/models/openai-oss.md`
}

func (m *GptOss120B) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *GptOss120B) GenerateContent(ctx context.Context, msgs ...*assistant.Message) (*assistant.GenerateContentResponse, error) {
	return generateContent(ctx, m.client, m.Name(), m.systemInstruction, msgs)
}

func (m *GptOss120B) GenerateContentStream(ctx context.Context, msgs ...*assistant.Message) (iter.Seq[*assistant.GenerateContentResponse], error) {
	return generateContentStream(ctx, m.client, m.Name(), m.systemInstruction, msgs)
}
