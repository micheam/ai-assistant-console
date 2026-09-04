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

type GPT56Luna struct {
	systemInstruction []*assistant.TextContent
	client            *APIClient
}

var _ assistant.GenerativeModel = (*GPT56Luna)(nil)

func NewGPT56Luna(apiKey string) *GPT56Luna {
	return &GPT56Luna{
		client: NewAPIClient(apiKey),
	}
}

func (m *GPT56Luna) Provider() string {
	return ProviderName
}

func (m *GPT56Luna) Name() string {
	return "gpt-5.6-luna"
}

func (m *GPT56Luna) Description() string {
	return `GPT-5.6 Luna is the fastest, most affordable model in the GPT-5.6 series.
It features a 1.05M context window and 128K max output tokens.
Pricing: $1.00 / $6.00 per MTok (input / output).
Reference: https://openai.com/index/gpt-5-6/`
}

func (m *GPT56Luna) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *GPT56Luna) SetHttpClient(c *http.Client) {
	m.client.SetHTTPClient(c)
}

func (m *GPT56Luna) GenerateContent(ctx context.Context, msgs ...assistant.Message) (*assistant.GenerateContentResponse, error) {
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

func (m *GPT56Luna) GenerateContentStream(ctx context.Context, msgs ...assistant.Message) (iter.Seq2[*assistant.GenerateContentResponse, error], error) {
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
