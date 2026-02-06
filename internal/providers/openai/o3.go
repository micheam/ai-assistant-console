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

type O3 struct {
	assistant.DeprecationInfo
	systemInstruction []*assistant.TextContent
	client            *APIClient
}

var _ assistant.GenerativeModel = (*O3)(nil)

func NewO3(apiKey string) *O3 {
	return &O3{
		client: NewAPIClient(apiKey),
	}
}

func (m *O3) Provider() string {
	return ProviderName
}

func (m *O3) Name() string {
	return "o3"
}

func (m *O3) Description() string {
	return `o3 is a powerful reasoning model that sets a new standard for math, science, coding, and visual reasoning tasks.
It features a 200K context window and 100K max output tokens, with a knowledge cutoff of June 2024.
It supports text and image inputs, structured outputs, and function calling.
Pricing: $0.40 / $1.60 per MTok (input / output).
Reference: https://platform.openai.com/docs/models#o3`
}

func (m *O3) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *O3) SetHttpClient(c *http.Client) {
	m.client.SetHTTPClient(c)
}

func (m *O3) GenerateContent(ctx context.Context, msgs ...*assistant.Message) (*assistant.GenerateContentResponse, error) {
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

func (m *O3) GenerateContentStream(ctx context.Context, msgs ...*assistant.Message) (iter.Seq[*assistant.GenerateContentResponse], error) {
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
