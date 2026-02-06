package cerebras

import (
	"context"
	"iter"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/providers/openai"
)

type Qwen3_235B_A22B_Instruct_2507 struct {
	assistant.DeprecationInfo
	systemInstruction []*assistant.TextContent
	client            *openai.APIClient
}

var _ assistant.GenerativeModel = (*Qwen3_235B_A22B_Instruct_2507)(nil)

func NewQwen3_235B_A22B_Instruct_2507(apiKey string) *Qwen3_235B_A22B_Instruct_2507 {
	return &Qwen3_235B_A22B_Instruct_2507{
		client: openai.NewAPIClient(apiKey),
	}
}

func (m *Qwen3_235B_A22B_Instruct_2507) Provider() string {
	return ProviderName
}

func (m *Qwen3_235B_A22B_Instruct_2507) Name() string {
	return "qwen-3-235b-a22b-instruct-2507"
}

func (m *Qwen3_235B_A22B_Instruct_2507) Description() string {
	return `Qwen 3 235B A22B Instruct 2507 - High-quality Preview model (~1400 tokens/sec).
Best for: Complex analysis, high-quality code generation, and detailed instruction following.
WARNING: This is a Preview model that may be discontinued on short notice.
Consider fallback to gpt-oss-120b for critical processes.
Reference: https://inference-docs.cerebras.ai/models/qwen-3-235b-2507.md`
}

func (m *Qwen3_235B_A22B_Instruct_2507) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *Qwen3_235B_A22B_Instruct_2507) GenerateContent(ctx context.Context, msgs ...*assistant.Message) (*assistant.GenerateContentResponse, error) {
	return openai.GenerateContent(ctx, m.client, Endpoint, m.Name(), m.systemInstruction, msgs)
}

func (m *Qwen3_235B_A22B_Instruct_2507) GenerateContentStream(ctx context.Context, msgs ...*assistant.Message) (iter.Seq[*assistant.GenerateContentResponse], error) {
	return openai.GenerateContentStream(ctx, m.client, Endpoint, m.Name(), m.systemInstruction, msgs)
}
