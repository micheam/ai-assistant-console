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

type GPT6Astra struct {
	systemInstruction []*assistant.TextContent
	client            *APIClient
}

var _ assistant.GenerativeModel = (*GPT6Astra)(nil)

func NewGPT6Astra(apiKey string) *GPT6Astra {
	return &GPT6Astra{
		client: NewAPIClient(apiKey),
	}
}

func (m *GPT6Astra) Provider() string {
	return ProviderName
}

func (m *GPT6Astra) Name() string {
	return "gpt-6-astra"
}

func (m *GPT6Astra) Description() string {
	return `GPT-6 Astra is OpenAI's flagship model, state-of-the-art on computer use, browsing,
software engineering, cybersecurity, science, and professional work.
It features a 1.05M context window and 128K max output tokens.
Pricing: $10.00 / $50.00 per MTok (input / output).
Reference: https://openai.com/index/gpt-6-astra/`
}

func (m *GPT6Astra) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *GPT6Astra) SetHttpClient(c *http.Client) {
	m.client.SetHTTPClient(c)
}

func (m *GPT6Astra) GenerateContent(ctx context.Context, msgs ...assistant.Message) (*assistant.GenerateContentResponse, error) {
	req, err := BuildChatRequest(ctx, m.Name(), m.systemInstruction, msgs)
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

func (m *GPT6Astra) GenerateContentStream(ctx context.Context, msgs ...assistant.Message) (iter.Seq2[*assistant.GenerateContentResponse, error], error) {
	req, err := BuildChatRequest(ctx, m.Name(), m.systemInstruction, msgs)
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
