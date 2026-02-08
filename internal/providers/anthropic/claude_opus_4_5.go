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

const ModelNameClaudeOpus4_5 = "claude-opus-4-5"

type ClaudeOpus4_5 struct {
	systemInstruction []*assistant.TextContent
	client            *anthropic.Client

	opts []anthropicopt.RequestOption
}

var _ assistant.GenerativeModel = (*ClaudeOpus4_5)(nil)

func NewClaudeOpus4_5(client *anthropic.Client) *ClaudeOpus4_5 { return &ClaudeOpus4_5{client: client} }
func (m *ClaudeOpus4_5) Provider() string                      { return ProviderName }
func (m *ClaudeOpus4_5) Name() string                          { return ModelNameClaudeOpus4_5 }
func (m *ClaudeOpus4_5) Description() string {
	return `[Deprecated] Claude Opus 4.5 - superseded by Claude Opus 4.6.
State-of-the-art on real-world software engineering tasks, scoring higher than human candidates.
Enhanced vision, reasoning, and mathematics skills. Best for long-horizon autonomous tasks,
tool calling, and multi-agent coordination. Supports 200K context window and 64K max output.`
}

func (m *ClaudeOpus4_5) SetSystemInstruction(contents ...*assistant.TextContent) {
	m.systemInstruction = contents
}

func (m *ClaudeOpus4_5) GenerateContent(
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

func (m *ClaudeOpus4_5) GenerateContentStream(
	ctx context.Context,
	msgs ...*assistant.Message,
) (iter.Seq2[*assistant.GenerateContentResponse, error], error) {
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
	return func(yield func(*assistant.GenerateContentResponse, error) bool) {
		for stream.Next() {
			event := stream.Current()
			message.Accumulate(event)

			switch delta := event.Delta.(type) {
			case anthropic.ContentBlockDeltaEventDelta:
				if delta.Text != "" {
					resp := &assistant.GenerateContentResponse{
						Content: assistant.NewTextContent(delta.Text),
					}
					if !yield(resp, nil) {
						return
					}
				}
			}
		}
		if err := stream.Err(); err != nil {
			logger.Error(fmt.Sprintf("stream error: %v", err))
			yield(nil, fmt.Errorf("anthropic stream error: %w", err))
		}
	}, nil
}
