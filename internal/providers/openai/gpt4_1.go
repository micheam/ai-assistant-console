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

type GPT41 struct {
	systemInstruction []*assistant.TextContent
	client            *APIClient
}

var _ assistant.GenerativeModel = (*GPT41)(nil)

func NewGPT41(apiKey string) *GPT41 {
	return &GPT41{
		client: NewAPIClient(apiKey),
	}
}

func (m *GPT41) Provider() string {
	return ProviderName
}

func (m *GPT41) Name() string {
	return "gpt-4.1"
}

func (m *GPT41) Description() string {
	return `GPT-4.1 is OpenAI's model optimized for coding tasks with a 1M token context window and 32K max output tokens.
It excels at instruction following and tool calling, with broad knowledge across domains.
26% less expensive than GPT-4o for median queries.
Pricing: $2.00 / $8.00 per MTok (input / output).
Reference: https://platform.openai.com/docs/models#gpt-4.1`
}

func (m *GPT41) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *GPT41) SetHttpClient(c *http.Client) {
	m.client.SetHTTPClient(c)
}

func (m *GPT41) GenerateContent(ctx context.Context, msgs ...*assistant.Message) (*assistant.GenerateContentResponse, error) {
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

func (m *GPT41) GenerateContentStream(ctx context.Context, msgs ...*assistant.Message) (iter.Seq[*assistant.GenerateContentResponse], error) {
	req, err := BuildChatRequest(ctx, m.Name(), m.systemInstruction, msgs)
	if err != nil {
		return nil, fmt.Errorf("build chat request: %w", err)
	}
	req.Stream = true
	iter, err := m.client.DoStream(ctx, endpoint, req)
	if err != nil {
		return nil, err
	}
	return func(yield func(*assistant.GenerateContentResponse) bool) {
		for s := range iter {
			var res *ChatResponse
			err := json.Unmarshal([]byte(s), &res)
			if err != nil {
				logging.LoggerFrom(ctx).Error(fmt.Sprintf("error: %v", err))
				continue
			}
			delta := assistant.NewTextContent(res.Choices[0].Delta.Content)
			if !yield(&assistant.GenerateContentResponse{Content: delta}) {
				break
			}
		}
	}, nil
}
