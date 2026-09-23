package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/logging"
)

type GPT6Sol struct {
	systemInstruction []*assistant.TextContent
	client            *APIClient
	genOpts           assistant.GenerationOptions
}

var (
	_ assistant.GenerativeModel         = (*GPT6Sol)(nil)
	_ assistant.GenerationOptionCapable = (*GPT6Sol)(nil)
)

func NewGPT6Sol(apiKey string) *GPT6Sol {
	return &GPT6Sol{
		client: NewAPIClient(apiKey),
	}
}

func (m *GPT6Sol) Provider() string {
	return ProviderName
}

func (m *GPT6Sol) Name() string {
	return "gpt-6-sol"
}

func (m *GPT6Sol) Description() string {
	return `GPT-6 Sol is a balanced model in the GPT-6 series, built for complex coding
and agentic workflows.
It features a 1.05M context window and 128K max output tokens.
Pricing: $2.00 / $10.00 per MTok (input / output).
Reference: https://openai.com/index/introducing-gpt-6-sol-and-luna/`
}

func (m *GPT6Sol) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *GPT6Sol) SetGenerationOptions(opts assistant.GenerationOptions) {
	m.genOpts = opts
}

func (m *GPT6Sol) SetHttpClient(c *http.Client) {
	m.client.SetHTTPClient(c)
}

func (m *GPT6Sol) GenerateContent(ctx context.Context, msgs ...assistant.Message) (*assistant.GenerateContentResponse, error) {
	req, err := BuildChatRequest(ctx, m.Name(), m.systemInstruction, m.genOpts, msgs)
	if err != nil {
		return nil, fmt.Errorf("build chat request: %w", err)
	}
	// Send request
	resp := new(ChatResponse)
	if err := m.client.DoPost(ctx, endpoint, req, resp); err != nil {
		return nil, err
	}
	return ToGenerateContentResponse(resp), nil
}

func (m *GPT6Sol) GenerateContentStream(ctx context.Context, msgs ...assistant.Message) (iter.Seq2[*assistant.GenerateContentResponse, error], error) {
	req, err := BuildChatRequest(ctx, m.Name(), m.systemInstruction, m.genOpts, msgs)
	if err != nil {
		return nil, fmt.Errorf("build chat request: %w", err)
	}
	// Send request
	req.Stream = true
	iter, err := m.client.DoStream(ctx, endpoint, req)
	if err != nil {
		return nil, err
	}
	// return converter iter
	return func(yield func(*assistant.GenerateContentResponse, error) bool) {
		for s := range iter {
			var res *ChatResponse
			err := json.Unmarshal([]byte(s), &res)
			if err != nil {
				logging.LoggerFrom(ctx).Error(fmt.Sprintf("unmarshal error: %v", err))
				yield(nil, fmt.Errorf("failed to unmarshal stream response: %w", err))
				continue
			}
			if len(res.Choices) == 0 || res.Choices[0].Delta == nil {
				continue
			}
			delta := assistant.NewTextContent(res.Choices[0].Delta.Content)
			if !yield(&assistant.GenerateContentResponse{Content: delta}, nil) {
				break
			}
		}
	}, nil
}
