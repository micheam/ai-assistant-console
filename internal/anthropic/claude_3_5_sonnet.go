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

type Claude3_5Sonnet struct {
	systemInstruction *assistant.TextContent
	client            *anthropic.Client

	opts []anthropicopt.RequestOption
}

var _ assistant.GenerativeModel = (*Claude3_5Sonnet)(nil)

func NewClaude3_5Sonnet(client *anthropic.Client) *Claude3_5Sonnet {
	return &Claude3_5Sonnet{client: client}
}

func (m *Claude3_5Sonnet) Name() string {
	return anthropic.ModelClaude3_5SonnetLatest
}
func (m *Claude3_5Sonnet) Description() string {
	return `Claude 3.5 Sonnet combines high-caliber intelligence with improved processing speed,
making it ideal for complex research, detailed analysis, and strategic planning.
Its balanced performance is tailored for advanced problem-solving and rapid execution in demanding tasks.`
}

func (m *Claude3_5Sonnet) SetSystemInstruction(text *assistant.TextContent) {
	m.systemInstruction = text
}

func (m *Claude3_5Sonnet) GenerateContent(
	ctx context.Context,
	msgs ...*assistant.Message,
) (*assistant.GenerateContentResponse, error) {
	logger := logging.LoggerFrom(ctx).With("provider", "anthropic", "model", m.Name())

	// Request to Anthropics API
	body, err := buildRequestBody(
		logging.ContextWith(ctx, logger),
		anthropic.ModelClaude3_5SonnetLatest,
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

func (m *Claude3_5Sonnet) GenerateContentStream(
	ctx context.Context,
	msgs ...*assistant.Message,
) (iter.Seq[*assistant.GenerateContentResponse], error) {
	logger := logging.LoggerFrom(ctx).With("provider", "anthropic", "model", m.Name())

	// Request to Anthropics API
	body, err := buildRequestBody(
		logging.ContextWith(ctx, logger),
		anthropic.ModelClaude3_5SonnetLatest,
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
