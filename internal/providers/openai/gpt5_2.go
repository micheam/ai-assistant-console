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

type GPT52 struct {
	systemInstruction []*assistant.TextContent
	client            *APIClient
}

var _ assistant.GenerativeModel = (*GPT52)(nil)

func NewGPT52(apiKey string) *GPT52 {
	return &GPT52{
		client: NewAPIClient(apiKey),
	}
}

func (m *GPT52) Provider() string {
	return ProviderName
}

func (m *GPT52) Name() string {
	return "gpt-5.2"
}

func (m *GPT52) Description() string {
	return `GPT-5.2 is OpenAI's flagship model for coding and agentic tasks.
It features a 400K context window and 128K max output tokens, with a knowledge cutoff of August 2025.
It excels at complex reasoning, coding (SWE-Bench Pro: 55.6%), math (AIME 2025: 100%), and science (GPQA Diamond: ~93%).
Pricing: $1.75 / $14.00 per MTok (input / output).
Reference: https://platform.openai.com/docs/models#gpt-5.2`
}

func (m *GPT52) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *GPT52) SetHttpClient(c *http.Client) {
	m.client.SetHTTPClient(c)
}

func (m *GPT52) GenerateContent(ctx context.Context, msgs ...*assistant.Message) (*assistant.GenerateContentResponse, error) {
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

func (m *GPT52) GenerateContentStream(ctx context.Context, msgs ...*assistant.Message) (iter.Seq[*assistant.GenerateContentResponse], error) {
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
