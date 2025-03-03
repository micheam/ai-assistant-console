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

type GPT4OMini struct {
	systemInstruction *assistant.TextContent
	client            *APIClient
}

var _ assistant.GenerativeModel = (*GPT4OMini)(nil)

func NewGPT4OMini(apiKey string) *GPT4OMini {
	return &GPT4OMini{
		client: NewAPIClient(apiKey),
	}
}

func (m *GPT4OMini) Name() string {
	return "gpt-4o-mini"
}

func (m *GPT4OMini) Description() string {
	return `GPT-4o-mini (“o” for “omni”) is a fast, affordable small model for focused tasks.
It accepts both text and image inputs, and produces text outputs, including structured outputs.
Outputs from a larger model like GPT-4o can be distilled into GPT-4o-mini to achieve similar
results with reduced cost and latency.
The knowledge cutoff date for GPT-4o-mini models is October 2023.
Reference: https://platform.openai.com/docs/models#gpt-4o-mini`
}

func (m *GPT4OMini) SetSystemInstruction(text *assistant.TextContent) {
	m.systemInstruction = text
}

func (m *GPT4OMini) SetHttpClient(c *http.Client) {
	m.client.SetHTTPClient(c)
}

func (m *GPT4OMini) GenerateContent(ctx context.Context, msgs ...*assistant.Message) (*assistant.GenerateContentResponse, error) {
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

func (m *GPT4OMini) GenerateContentStream(ctx context.Context, msgs ...*assistant.Message) (iter.Seq[*assistant.GenerateContentResponse], error) {
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
