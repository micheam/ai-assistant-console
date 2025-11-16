package anthropic

import (
	"context"
	"fmt"
	"iter"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	anthropicopt "github.com/anthropics/anthropic-sdk-go/option"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/logging"
)

type ClaudeHaiku4_5 struct {
	systemInstruction *assistant.TextContent
	client            *anthropic.Client

	opts []anthropicopt.RequestOption
}

var _ assistant.GenerativeModel = (*ClaudeHaiku4_5)(nil)

func NewClaudeHaiku4_5(client *anthropic.Client) *ClaudeHaiku4_5 {
	return &ClaudeHaiku4_5{client: client}
}

func (m *ClaudeHaiku4_5) Name() string {
	return "claude-haiku-4-5"
}
func (m *ClaudeHaiku4_5) Description() string {
	return `Claude Haiku 4.5 is the fastest model with near-frontier performance.
Engineered for lightning-fast speed at the most economical price point.
Best for real-time applications, high-volume intelligent processing,
cost-sensitive deployments needing strong reasoning, and sub-agent tasks.
Supports 200K context window.`
}

func (m *ClaudeHaiku4_5) SetSystemInstruction(text *assistant.TextContent) {
	m.systemInstruction = text
}

func (m *ClaudeHaiku4_5) GenerateContent(
	ctx context.Context,
	msgs ...*assistant.Message,
) (*assistant.GenerateContentResponse, error) {
	logger := logging.LoggerFrom(ctx).With("provider", "anthropic", "model", m.Name())

	// Request to Anthropics API
	body, err := buildRequestBody(
		logging.ContextWith(ctx, logger),
		anthropic.Model(m.Name()),
		m.systemInstruction,
		msgs)
	if err != nil {
		return nil, fmt.Errorf("anthropic request body: %w", err)
	}
	res, err := m.client.Messages.New(ctx, *body, m.opts...)
	if err != nil {
		return nil, fmt.Errorf("anthropic New Message: %w", err)
	}

	// Handle Response
	logger = logger.With("request-id", res.ID)
	if res.StopReason != "end_turn" {
		logger.Warn(fmt.Sprintf("anthropic response stop with reason: %s", res.StopReason))
	}
	if len(res.Content) > 1 {
		logger.Warn("anthropic response has more than one content", "content", fmt.Sprintf("%+v", res.Content))
	}
	return &assistant.GenerateContentResponse{
		Content: assistant.NewTextContent(res.Content[0].Text),
	}, nil
}

func (m *ClaudeHaiku4_5) GenerateContentStream(
	ctx context.Context,
	msgs ...*assistant.Message,
) (iter.Seq[*assistant.GenerateContentResponse], error) {
	logger := logging.LoggerFrom(ctx).With("provider", "anthropic", "model", m.Name())

	// Request to Anthropics API
	body, err := buildRequestBody(
		logging.ContextWith(ctx, logger),
		anthropic.Model(m.Name()),
		m.systemInstruction,
		msgs)
	if err != nil {
		return nil, fmt.Errorf("anthropic request body: %w", err)
	}
	stream := m.client.Messages.NewStreaming(ctx, *body, m.opts...)

	// return converter iter
	message := anthropic.Message{}
	return func(yield func(*assistant.GenerateContentResponse) bool) {
		for stream.Next() {
			event := stream.Current()
			message.Accumulate(event)

			switch delta := event.Delta.(type) {
			case anthropic.ContentBlockDeltaEventDelta:
				if delta.Text != "" {
					resp := &assistant.GenerateContentResponse{
						Content: assistant.NewTextContent(delta.Text),
					}
					if !yield(resp) {
						return
					}
				}
			}
		}
		if err := stream.Err(); err != nil {
			logger.Error(fmt.Sprintf("error: %v", err))
		}
	}, nil
}
