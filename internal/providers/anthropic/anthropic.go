package anthropic

import (
	"context"
	"fmt"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/logging"
)

const (
	ProviderName     = "anthropic"
	defaultMaxTokens = 1_024 * 8
)

// Anthropic available models and their descriptions from Anthropic Documentation:
// * https://platform.claude.com/docs/en/about-claude/models/overview
//
// Claude Opus 4.6:
//
//     * Our most intelligent model for building agents and coding
//     * Top-tier results in reasoning, coding, multilingual tasks, and long-context handling
//     * Supports extended thinking and adaptive thinking
//     * Pricing: $5/MTok input, $25/MTok output
//     * Supports 200K context window (1M with beta header) and 128K max output
//
// Claude Sonnet 4.5:
//
//     * Our best combination of speed and intelligence
//     * Best for complex agents and coding with superior tool orchestration
//     * Ideal for autonomous coding agents, complex financial analysis, multi-hour research tasks
//     * Pricing: $3/MTok input, $15/MTok output
//     * Supports 200K context window (1M with beta header) and 64K max output
//
// Claude Haiku 4.5:
//
//     * Our fastest model with near-frontier intelligence
//     * Most economical price point with lightning-fast speed
//     * Best for real-time applications, high-volume intelligent processing, sub-agent tasks
//     * Pricing: $1/MTok input, $5/MTok output
//     * Supports 200K context window and 64K max output
//
// Claude Opus 4.5 (Deprecated):
//
//     * State-of-the-art model for the world's hardest problems (legacy)
//     * Superseded by Claude Opus 4.6
//     * Pricing: $5/MTok input, $25/MTok output
//     * Supports 200K context window and 64K max output

// AvailableModels returns a list of available models
func AvailableModels() []assistant.ModelDescriptor {
	return []assistant.ModelDescriptor{
		&ClaudeOpus4_6{},
		&ClaudeSonnet4_5{},
		&ClaudeOpus4_5{DeprecationInfo: assistant.DeprecationInfo{IsDeprecated: true, RemovedIn: "v2.0.0"}},
		&ClaudeHaiku4_5{},
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
	case "claude-opus-4-6":
		return &ClaudeOpus4_6{}, true
	case "claude-sonnet-4-5":
		return &ClaudeSonnet4_5{}, true
	case "claude-opus-4-5":
		return &ClaudeOpus4_5{}, true
	case "claude-haiku-4-5":
		return &ClaudeHaiku4_5{}, true
	}
}

// NewGenerativeModel creates a new instance of a generative model
func NewGenerativeModel(modelName, apiKey string) (assistant.GenerativeModel, error) {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	switch modelName {
	case "claude-opus-4-6":
		return NewClaudeOpus4_6(client), nil
	case "claude-sonnet-4-5":
		return NewClaudeSonnet4(client), nil
	case "claude-opus-4-5":
		return NewClaudeOpus4_5(client), nil
	case "claude-haiku-4-5":
		return NewClaudeHaiku4_5(client), nil
	}
	return nil, fmt.Errorf("unsupported model name: %s", modelName)
}

func buildRequestBody(ctx context.Context, model anthropic.Model, systemInstruction []*assistant.TextContent, msgs []*assistant.Message) (*anthropic.MessageNewParams, error) {
	messages, err := messageParams(ctx, msgs...)
	if err != nil {
		return nil, fmt.Errorf("build message params: %w", err)
	}
	return &anthropic.MessageNewParams{
		MaxTokens: anthropic.F(int64(defaultMaxTokens)),
		Model:     anthropic.F(model),
		Messages:  anthropic.F(messages),
		System:    anthropic.F(systemMessageParam(systemInstruction)),
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
		var msg anthropic.MessageParam
		switch v := src.GetContents()[0].(type) {
		case *assistant.TextContent:
			msg = anthropic.NewAssistantMessage(anthropic.NewTextBlock(v.Text))
		case *assistant.AttachmentContent:
			msg = anthropic.NewAssistantMessage(anthropic.NewTextBlock(v.ToText()))
		default:
			return nil, fmt.Errorf("unsupported assistant message content type: %T", v)
		}
		return &msg, nil

	case assistant.MessageAuthorUser:
		contents := []anthropic.ContentBlockParamUnion{}
		for _, content := range src.Contents {
			switch c := content.(type) {
			case *assistant.TextContent:
				contents = append(contents, anthropic.NewTextBlock(c.Text))
			case *assistant.AttachmentContent:
				contents = append(contents, anthropic.NewTextBlock(c.ToText()))
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

func systemMessageParam(conts []*assistant.TextContent) []anthropic.TextBlockParam {
	if conts == nil {
		return []anthropic.TextBlockParam{}
	}
	param := make([]anthropic.TextBlockParam, 0, len(conts))
	for _, conts := range conts {
		param = append(param, anthropic.NewTextBlock(conts.Text))
	}
	return param
}
