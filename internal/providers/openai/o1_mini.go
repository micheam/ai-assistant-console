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

type O1Mini struct {
	systemInstruction []*assistant.TextContent
	client            *APIClient
}

var _ assistant.GenerativeModel = (*O1Mini)(nil)

func NewO1Mini(apiKey string) *O1Mini {
	return &O1Mini{
		client: NewAPIClient(apiKey),
	}
}

func (m *O1Mini) Provider() string {
	return ProviderName
}

func (m *O1Mini) Name() string {
	return "o1-mini"
}

func (m *O1Mini) Description() string {
	return `[Deprecated] o1-mini - superseded by o4-mini.
The o1 series of models are trained with reinforcement learning for complex reasoning tasks.
These models generate a long internal chain of thought before responding.
The knowledge cutoff date for o1-mini models is October 2023.
Reference: https://platform.openai.com/docs/models#o1`
}

func (m *O1Mini) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *O1Mini) SetHttpClient(c *http.Client) {
	m.client.SetHTTPClient(c)
}

func (m *O1Mini) GenerateContent(ctx context.Context, msgs ...*assistant.Message) (*assistant.GenerateContentResponse, error) {
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

func (m *O1Mini) GenerateContentStream(ctx context.Context, msgs ...*assistant.Message) (iter.Seq[*assistant.GenerateContentResponse], error) {
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
			if len(res.Choices) == 0 || res.Choices[0].Delta == nil {
				continue
			}
			delta := assistant.NewTextContent(res.Choices[0].Delta.Content)
			if !yield(&assistant.GenerateContentResponse{Content: delta}) {
				break
			}
		}
	}, nil
}
