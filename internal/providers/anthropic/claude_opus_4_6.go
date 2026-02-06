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

const ModelNameClaudeOpus4_6 = "claude-opus-4-6"

type ClaudeOpus4_6 struct {
	systemInstruction []*assistant.TextContent
	client            *anthropic.Client

	opts []anthropicopt.RequestOption
}

var _ assistant.GenerativeModel = (*ClaudeOpus4_6)(nil)

func NewClaudeOpus4_6(client *anthropic.Client) *ClaudeOpus4_6 { return &ClaudeOpus4_6{client: client} }
func (m *ClaudeOpus4_6) Provider() string                      { return ProviderName }
func (m *ClaudeOpus4_6) Name() string                          { return ModelNameClaudeOpus4_6 }
func (m *ClaudeOpus4_6) Description() string {
	return `Claude Opus 4.6 is the most intelligent model for building agents and coding.
Top-tier results in reasoning, coding, multilingual tasks, and long-context handling.
Supports extended thinking and adaptive thinking. Pricing: $5/MTok input, $25/MTok output.
Supports 200K context window (1M with beta header) and 128K max output.`
}

func (m *ClaudeOpus4_6) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *ClaudeOpus4_6) GenerateContent(
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
	if len(res.Content) == 0 {
		return nil, fmt.Errorf("anthropic response has no content")
	}
	if len(res.Content) > 1 {
		logger.Warn("anthropic response has more than one content", "content", fmt.Sprintf("%+v", res.Content))
	}
	return &assistant.GenerateContentResponse{
		Content: assistant.NewTextContent(res.Content[0].Text),
	}, nil
}

func (m *ClaudeOpus4_6) GenerateContentStream(
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
