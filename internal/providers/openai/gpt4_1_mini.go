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

type GPT41Mini struct {
	systemInstruction []*assistant.TextContent
	client            *APIClient
}

var _ assistant.GenerativeModel = (*GPT41Mini)(nil)

func NewGPT41Mini(apiKey string) *GPT41Mini {
	return &GPT41Mini{
		client: NewAPIClient(apiKey),
	}
}

func (m *GPT41Mini) Provider() string {
	return ProviderName
}

func (m *GPT41Mini) Name() string {
	return "gpt-4.1-mini"
}

func (m *GPT41Mini) Description() string {
	return `GPT-4.1 mini is a fast, capable, and efficient small model with a 1M token context window and 32K max output tokens.
It excels in instruction-following, coding, and overall intelligence at reduced cost and latency.
Pricing: $0.40 / $1.60 per MTok (input / output).
Reference: https://platform.openai.com/docs/models#gpt-4.1-mini`
}

func (m *GPT41Mini) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *GPT41Mini) SetHttpClient(c *http.Client) {
	m.client.SetHTTPClient(c)
}

func (m *GPT41Mini) GenerateContent(ctx context.Context, msgs ...*assistant.Message) (*assistant.GenerateContentResponse, error) {
	req, err := BuildChatRequest(ctx, m.Name(), m.systemInstruction, msgs)
	if err != nil {
		return nil, fmt.Errorf("build chat request: %w", err)
	}
	resp := new(ChatResponse)
	if err := m.client.DoPost(ctx, endpoint, req, resp); err != nil {
		return nil, err
	}
	return ToGenerateContentResponse(resp), nil
}

func (m *GPT41Mini) GenerateContentStream(ctx context.Context, msgs ...*assistant.Message) (iter.Seq2[*assistant.GenerateContentResponse, error], error) {
	req, err := BuildChatRequest(ctx, m.Name(), m.systemInstruction, msgs)
	if err != nil {
		return nil, fmt.Errorf("build chat request: %w", err)
	}
	req.Stream = true
	iter, err := m.client.DoStream(ctx, endpoint, req)
	if err != nil {
		return nil, err
	}
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
