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

type O3Mini struct {
	assistant.DeprecationInfo
	systemInstruction []*assistant.TextContent
	client            *APIClient
}

var _ assistant.GenerativeModel = (*O3Mini)(nil)

func NewO3Mini(apiKey string) *O3Mini {
	return &O3Mini{
		client: NewAPIClient(apiKey),
	}
}

func (m *O3Mini) Provider() string {
	return ProviderName
}

func (m *O3Mini) Name() string {
	return "o3-mini"
}

func (m *O3Mini) Description() string {
	return `The o3 series of models are trained with reinforcement learning for complex reasoning tasks.
These models generate a long internal chain of thought before responding.
o3-mini is the most recent small reasoning model, providing high intelligence at the same cost and latency as o1-mini.
It supports key developer features, including structured outputs, function calling, and the Batch API.
Like other models in the o-series, it is optimised for science, math, and coding tasks.
The knowledge cutoff date for o3-mini models is October 2023.
Reference: https://platform.openai.com/docs/models#o3-mini`
}

func (m *O3Mini) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *O3Mini) SetHttpClient(c *http.Client) {
	m.client.SetHTTPClient(c)
}

func (m *O3Mini) GenerateContent(ctx context.Context, msgs ...*assistant.Message) (*assistant.GenerateContentResponse, error) {
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

func (m *O3Mini) GenerateContentStream(ctx context.Context, msgs ...*assistant.Message) (iter.Seq[*assistant.GenerateContentResponse], error) {
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
