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

type O1 struct {
	systemInstruction *assistant.TextContent
	client            *APIClient
}

var _ assistant.GenerativeModel = (*O1)(nil)

func NewO1(apiKey string) *O1 {
	return &O1{
		client: NewAPIClient(apiKey),
	}
}

func (m *O1) Name() string {
	return "o1"
}

func (m *O1) Description() string {
	return `The o1 series of models are trained with reinforcement learning for complex reasoning tasks.
These models generate a long internal chain of thought before responding.
o1 is the latest reasoning model, supporting both text and image inputs.
The knowledge cutoff date for o1 models is October 2023.
Reference: https://platform.openai.com/docs/models#o1`
}

func (m *O1) SetSystemInstruction(text *assistant.TextContent) {
	m.systemInstruction = text
}

func (m *O1) SetHttpClient(c *http.Client) {
	m.client.SetHTTPClient(c)
}

func (m *O1) GenerateContent(ctx context.Context, msgs ...*assistant.Message) (*assistant.GenerateContentResponse, error) {
	req, err := buildChatRequest(ctx, m.Name(), m.systemInstruction, msgs)
	if err != nil {
		return nil, fmt.Errorf("build chat request: %w", err)
	}
	// Send request
	resp := new(ChatResponse)
	if err := m.client.DoPost(ctx, endpoint, req, resp); err != nil {
		return nil, err
	}
	return toGenerateContentResponse(resp), nil
}

func (m *O1) GenerateContentStream(ctx context.Context, msgs ...*assistant.Message) (iter.Seq[*assistant.GenerateContentResponse], error) {
	req, err := buildChatRequest(ctx, m.Name(), m.systemInstruction, msgs)
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
