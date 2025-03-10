package anthropic

import (
	"context"
	"fmt"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/logging"
)

// Anthropic available models and their descriptions.
// 2025-03-10 16:00 JST
//
// Claude 3.7 Sonnet:
//
//     * Most intelligent model with extended thinking capabilities
//     * Ideal for complex reasoning, advanced problem-solving, and strategic analysis
//     * Shows thinking process before delivering answers 1
//
// Claude 3.5 Family:
// Sonnet:
//
//     * Most intelligent model combining top performance with improved speed
//     * Best for advanced research/analysis, complex problem-solving, and strategic planning 2
//
// Haiku:
//
//     * Fastest and most cost-effective model
//     * Excels at code generation, real-time chatbots, data extraction and labeling , 3
//
// Claude 3 Family:
//
// Opus:
//
//     * Best performance on complex tasks like math and coding
//     * Great for task automation, R&D, strategic analysis 4
//
// Sonnet:
//
//     * Balances intelligence and speed
//     * Ideal for data processing, sales forecasting, code generation
//
// Haiku:
//
//     * Near-instant responsiveness
//     * Perfect for live support chat, translations, content moderation
//
// Each model offers different context window sizes and pricing options to match specific use case needs.
// For example, Claude 3.7 Sonnet has a 200k token context window and extended thinking capabilities 1,
// while Claude 3.5 Haiku focuses on fast, cost-effective performance starting at $0.80 per million input tokens 3.

// AvailableModels returns a list of available models
func AvailableModels() []string {
	return []string{
		"claude-3-7-sonnet",
	}
}

func describeModel(modelName string) (desc string, found bool) {
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
	case "claude-3-7-sonnet":
		return &Claude3_7Sonnet{}, true
	}
}

// NewGenerativeModel creates a new instance of a generative model
func NewGenerativeModel(modelName, apiKey string) (assistant.GenerativeModel, error) {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	switch modelName {
	case "claude-3-7-sonnet":
		return NewClaude3_7Sonnet(client), nil
	}
	return nil, fmt.Errorf("unsupported model name: %s", modelName)
}

func buildRequestBody(ctx context.Context, model anthropic.Model, smsg *assistant.TextContent, msgs []*assistant.Message) (*anthropic.MessageNewParams, error) {
	messages, err := messageParams(ctx, msgs...)
	if err != nil {
		return nil, fmt.Errorf("build message params: %w", err)
	}
	return &anthropic.MessageNewParams{
		MaxTokens: anthropic.F(int64(1_024)),
		Model:     anthropic.F(model),
		Messages:  anthropic.F(messages),
		System:    anthropic.F(systemMessageParam(smsg)),
	}, nil
}

func messageParamFrom(ctx context.Context, src assistant.Message) (*anthropic.MessageParam, error) {
	logger := logging.LoggerFrom(ctx)

	switch src.Author {
	default:
		return nil, fmt.Errorf("unknown author: %s", src.Author)

	case assistant.MessageAuthorAssistant:
		if len(src.Contents) == 0 {
			return nil, fmt.Errorf("empty assistant message")
		}
		textContent, ok := src.Contents[0].(*assistant.TextContent)
		if !ok {
			return nil, fmt.Errorf("unexpected assistant message content: %T", src.Contents[0])
		}
		msg := anthropic.NewAssistantMessage(anthropic.NewTextBlock(textContent.Text))
		return &msg, nil

	case assistant.MessageAuthorUser:
		contents := []anthropic.ContentBlockParamUnion{}
		for _, content := range src.Contents {
			switch c := content.(type) {
			case *assistant.TextContent:
				contents = append(contents, anthropic.NewTextBlock(c.Text))
			default:
				logger.Warn("ignore unsupported content type", "type", fmt.Sprintf("%T", c))
			}
		}
		msg := anthropic.NewUserMessage(contents...)
		return &msg, nil
	}
}

func messageParams(ctx context.Context, msgs ...*assistant.Message) ([]anthropic.MessageParam, error) {
	var messages []anthropic.MessageParam
	for _, msg := range msgs {
		m, err := messageParamFrom(ctx, *msg)
		if err != nil {
			return nil, fmt.Errorf("anthropic message param from: %w", err)
		}
		messages = append(messages, *m)
	}
	return messages, nil
}

func systemMessageParam(text *assistant.TextContent) []anthropic.TextBlockParam {
	if text == nil {
		return []anthropic.TextBlockParam{}
	}
	return []anthropic.TextBlockParam{
		anthropic.NewTextBlock(text.Text),
	}
}
