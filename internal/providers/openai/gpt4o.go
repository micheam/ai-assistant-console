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

type GPT4O struct {
	systemInstruction []*assistant.TextContent
	client            *APIClient
}

var _ assistant.GenerativeModel = (*GPT4O)(nil)

func NewGPT4O(apiKey string) *GPT4O {
	return &GPT4O{
		client: NewAPIClient(apiKey),
	}
}

func (m *GPT4O) Provider() string {
	return ProviderName
}

func (m *GPT4O) Name() string {
	return "gpt-4o"
}

func (m *GPT4O) Description() string {
	return `[Deprecated] GPT-4o - superseded by gpt-5.2 and gpt-4.1.
GPT-4o ("o" for "omni") is a versatile, high-intelligence flagship model of OpenAI.
It supports both text and image inputs and generates text outputs, including structured responses.
The knowledge cutoff date for GPT-4o models is October 2023.
Reference: https://platform.openai.com/docs/models#gpt-4o`
}

func (m *GPT4O) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *GPT4O) SetHttpClient(c *http.Client) {
	m.client.SetHTTPClient(c)
}

func (m *GPT4O) GenerateContent(ctx context.Context, msgs ...*assistant.Message) (*assistant.GenerateContentResponse, error) {
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

func (m *GPT4O) GenerateContentStream(ctx context.Context, msgs ...*assistant.Message) (iter.Seq2[*assistant.GenerateContentResponse, error], error) {
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
