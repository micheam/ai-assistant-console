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

type O4Mini struct {
	systemInstruction []*assistant.TextContent
	client            *APIClient
}

var _ assistant.GenerativeModel = (*O4Mini)(nil)

func NewO4Mini(apiKey string) *O4Mini {
	return &O4Mini{
		client: NewAPIClient(apiKey),
	}
}

func (m *O4Mini) Provider() string {
	return ProviderName
}

func (m *O4Mini) Name() string {
	return "o4-mini"
}

func (m *O4Mini) Description() string {
	return `o4-mini is optimized for fast, effective reasoning with efficient performance in coding and visual tasks.
It features a 200K context window and 100K max output tokens, with a knowledge cutoff of May 2024.
It is faster and more affordable than o3, supporting text and image inputs.
Pricing: $1.10 / $4.40 per MTok (input / output).
Reference: https://platform.openai.com/docs/models#o4-mini`
}

func (m *O4Mini) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *O4Mini) SetHttpClient(c *http.Client) {
	m.client.SetHTTPClient(c)
}

func (m *O4Mini) GenerateContent(ctx context.Context, msgs ...*assistant.Message) (*assistant.GenerateContentResponse, error) {
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

func (m *O4Mini) GenerateContentStream(ctx context.Context, msgs ...*assistant.Message) (iter.Seq[*assistant.GenerateContentResponse], error) {
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
