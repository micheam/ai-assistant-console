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

type Claude3_5Haiku struct {
	systemInstruction *assistant.TextContent
	client            *anthropic.Client

	opts []anthropicopt.RequestOption
}

var _ assistant.GenerativeModel = (*Claude3_5Haiku)(nil)

func NewClaude3_5Haiku(client *anthropic.Client) *Claude3_5Haiku {
	return &Claude3_5Haiku{client: client}
}

func (m *Claude3_5Haiku) Name() string {
	return anthropic.ModelClaude3_5HaikuLatest
}
func (m *Claude3_5Haiku) Description() string {
	return `Claude 3.5 Haiku is engineered for speed and cost-efficiency,
delivering near-instant responses ideal for real-time applications.
It excels in rapid code generation, dynamic chatbot interactions, and
data extraction tasks, providing a balanced solution where quick
turnaround is essential.`
}

func (m *Claude3_5Haiku) SetSystemInstruction(text *assistant.TextContent) {
	m.systemInstruction = text
}

func (m *Claude3_5Haiku) GenerateContent(
	ctx context.Context,
	msgs ...*assistant.Message,
) (*assistant.GenerateContentResponse, error) {
	logger := logging.LoggerFrom(ctx).With("provider", "anthropic", "model", m.Name())

	// Request to Anthropics API
	body, err := buildRequestBody(
		logging.ContextWith(ctx, logger),
		anthropic.ModelClaude3_5HaikuLatest,
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

func (m *Claude3_5Haiku) GenerateContentStream(
	ctx context.Context,
	msgs ...*assistant.Message,
) (iter.Seq[*assistant.GenerateContentResponse], error) {
	logger := logging.LoggerFrom(ctx).With("provider", "anthropic", "model", m.Name())

	// Request to Anthropics API
	body, err := buildRequestBody(
		logging.ContextWith(ctx, logger),
		anthropic.ModelClaude3_5HaikuLatest,
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
