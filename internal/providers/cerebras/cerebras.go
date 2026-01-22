package cerebras

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/logging"
	"micheam.com/aico/internal/providers/openai"
)

const endpoint = "https://api.cerebras.ai/v1/chat/completions"
const ProviderName = "cerebras"

// AvailableModels returns a list of available models
func AvailableModels() []assistant.ModelDescriptor {
	return []assistant.ModelDescriptor{
		&Llama3_3_70B{},
		&Llama3_1_8B{},
	}
}

func DescribeModel(modelName string) (desc string, found bool) {
	m, ok := selectModel(modelName)
	if !ok {
		return "", false
	}
	return m.Description(), true
}

func selectModel(modelName string) (assistant.GenerativeModel, bool) {
	switch modelName {
	default:
		return nil, false
	case "llama-3.3-70b":
		return &Llama3_3_70B{}, true
	case "llama-3.1-8b":
		return &Llama3_1_8B{}, true
	}
}

// NewGenerativeModel creates a new instance of a generative model
func NewGenerativeModel(modelName, apiKey string) (assistant.GenerativeModel, error) {
	switch modelName {
	case "llama-3.3-70b":
		return NewLlama3_3_70B(apiKey), nil
	case "llama-3.1-8b":
		return NewLlama3_1_8B(apiKey), nil
	}
	return nil, fmt.Errorf("unsupported model name: %s", modelName)
}

// buildChatRequest builds a chat request for Cerebras API (OpenAI-compatible)
func buildChatRequest(ctx context.Context, modelName string, systemInstruction []*assistant.TextContent, msgs []*assistant.Message) (*openai.ChatRequest, error) {
	if len(msgs) == 0 {
		return nil, fmt.Errorf("no messages provided")
	}
	req := &openai.ChatRequest{
		Model:    modelName,
		Messages: make([]openai.Message, 0, len(msgs)+1),
	}
	if len(systemInstruction) > 0 {
		for _, c := range systemInstruction {
			req.Messages = append(req.Messages, &openai.SystemMessage{Content: c.Text})
		}
	}
	for _, msg := range msgs {
		for _, content := range msg.Contents {
			switch v := content.(type) {
			case *assistant.AttachmentContent:
				content := []openai.Content{&openai.TextContent{
					Text: v.ToText(),
				}}
				switch msg.Author {
				case "user":
					req.Messages = append(req.Messages, &openai.UserMessage{Content: content})
				case "assistant":
					req.Messages = append(req.Messages, &openai.AssistantMessage{Content: content})
				}

			case *assistant.TextContent:
				switch msg.Author {
				case "user":
					req.Messages = append(req.Messages, &openai.UserMessage{
						Content: []openai.Content{&openai.TextContent{Text: v.Text}},
					})
				case "assistant":
					req.Messages = append(req.Messages, &openai.AssistantMessage{
						Content: []openai.Content{&openai.TextContent{Text: v.Text}},
					})
				}

			case *assistant.URLImageContent:
				switch msg.Author {
				case "user":
					req.Messages = append(req.Messages, &openai.UserMessage{
						Content: []openai.Content{&openai.ImageContent{URL: v.URL}},
					})
				case "assistant":
					req.Messages = append(req.Messages, &openai.AssistantMessage{
						Content: []openai.Content{&openai.ImageContent{URL: v.URL}},
					})
				}

			default:
				logging.LoggerFrom(ctx).Warn(fmt.Sprintf("Unsupported message content type: %T", v))
			}
		}
	}
	return req, nil
}

func toGenerateContentResponse(src *openai.ChatResponse) *assistant.GenerateContentResponse {
	if len(src.Choices) > 0 {
		text := src.Choices[0].Message.Content[0].(*openai.TextContent).Text
		return &assistant.GenerateContentResponse{Content: &assistant.TextContent{Text: text}}
	}
	return &assistant.GenerateContentResponse{Content: nil}
}

// generateContent is a shared implementation for generating content
func generateContent(ctx context.Context, client *openai.APIClient, modelName string, systemInstruction []*assistant.TextContent, msgs []*assistant.Message) (*assistant.GenerateContentResponse, error) {
	req, err := buildChatRequest(ctx, modelName, systemInstruction, msgs)
	if err != nil {
		return nil, fmt.Errorf("build chat request: %w", err)
	}
	resp := new(openai.ChatResponse)
	if err := client.DoPost(ctx, endpoint, req, resp); err != nil {
		return nil, err
	}
	return toGenerateContentResponse(resp), nil
}

// generateContentStream is a shared implementation for streaming content
func generateContentStream(ctx context.Context, client *openai.APIClient, modelName string, systemInstruction []*assistant.TextContent, msgs []*assistant.Message) (iter.Seq[*assistant.GenerateContentResponse], error) {
	req, err := buildChatRequest(ctx, modelName, systemInstruction, msgs)
	if err != nil {
		return nil, fmt.Errorf("build chat request: %w", err)
	}
	req.Stream = true
	iter, err := client.DoStream(ctx, endpoint, req)
	if err != nil {
		return nil, err
	}
	return func(yield func(*assistant.GenerateContentResponse) bool) {
		for s := range iter {
			var res *openai.ChatResponse
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
