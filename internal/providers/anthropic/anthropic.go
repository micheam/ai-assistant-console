package anthropic

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/logging"
)

const (
	ProviderName = "anthropic"

	// defaultMaxTokens must be large enough that a propose_edit tool_use
	// input (which can carry a sizeable old_string/new_string pair) isn't
	// routinely cut off by the max_tokens limit; a truncated tool_use input
	// fails the turn (see streamContent).
	defaultMaxTokens = 1_024 * 32
)

// Anthropic available models and their descriptions from Anthropic Documentation:
// * https://platform.claude.com/docs/en/about-claude/models/overview
//
// Claude Fable 5.1:
//
//     * For demanding reasoning and long-horizon agentic work
//     * Successor to Claude Fable 5
//     * Supports adaptive thinking (always on) and effort control
//     * Pricing: $10/MTok input, $50/MTok output
//     * Supports 1M context window and 128K max output
//
// Claude Fable 5 (Legacy):
//
//     * Our most powerful model, excelling at creative, agentic, and coding tasks
//     * Supports adaptive thinking and effort control
//     * Pricing: $10/MTok input, $50/MTok output
//     * Supports 1M context window and 128K max output
//
// Claude Opus 5.5:
//
//     * The latest Opus model for long-running agentic coding and knowledge work
//     * Successor to Claude Opus 5
//     * Supports adaptive thinking (always on) and effort control
//     * Pricing: $4/MTok input, $20/MTok output
//     * Supports 1M context window and 128K max output
//
// Claude Opus 5:
//
//     * The Opus model for long-running agentic coding and knowledge work
//     * Successor to Claude Opus 4.8
//     * Supports adaptive thinking (on by default) and effort control
//     * Pricing: $5/MTok input, $25/MTok output
//     * Supports 1M context window and 128K max output
//
// Claude Opus 4.8:
//
//     * The latest Opus model for building agents and coding
//     * Top-tier results in reasoning, coding, multilingual tasks, and long-context handling
//     * Supports adaptive thinking and effort control
//     * Pricing: $5/MTok input, $25/MTok output
//     * Supports 1M context window and 128K max output
//
// Claude Sonnet 5:
//
//     * The best combination of speed and intelligence
//     * Successor to Claude Sonnet 4.6
//     * Supports adaptive thinking and effort control
//     * Pricing: $3/MTok input, $15/MTok output
//     * Supports 1M context window and 128K max output
//
// Claude Sonnet 4.6:
//
//     * The best combination of speed and intelligence
//     * Supports extended thinking and adaptive thinking
//     * Fast comparative latency
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
// Claude Opus 4.6 (Deprecated):
//
//     * Superseded by Claude Opus 4.8
//     * Top-tier results in reasoning, coding, multilingual tasks, and long-context handling
//     * Supports extended thinking and adaptive thinking
//     * Pricing: $5/MTok input, $25/MTok output
//     * Supports 200K context window (1M with beta header) and 128K max output

// AvailableModels returns a list of available models
func AvailableModels() []assistant.ModelDescriptor {
	return []assistant.ModelDescriptor{
		&ClaudeFable5_1{},
		&ClaudeFable5{},
		&ClaudeOpus5_5{},
		&ClaudeOpus5{},
		&ClaudeOpus4_8{},
		&ClaudeOpus4_6{},
		&ClaudeSonnet5{},
		&ClaudeSonnet4_6{},
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

// aliases maps a short family name to the latest supported model of that
// family. Every target must be a non-deprecated entry of AvailableModels.
var aliases = map[string]string{
	"fable":  ModelNameClaudeFable5_1,
	"opus":   ModelNameClaudeOpus5_5,
	"sonnet": ModelNameClaudeSonnet5,
	"haiku":  ModelNameClaudeHaiku4_5,
}

// Aliases returns the alias-to-model-name table.
func Aliases() map[string]string { return aliases }

func selectModel(modelName string) (assistant.GenerativeModel, bool) {
	switch modelName {
	default:
		return nil, false
	case "claude-fable-5-1":
		return &ClaudeFable5_1{}, true
	case "claude-fable-5":
		return &ClaudeFable5{}, true
	case "claude-opus-5-5":
		return &ClaudeOpus5_5{}, true
	case "claude-opus-5":
		return &ClaudeOpus5{}, true
	case "claude-opus-4-8":
		return &ClaudeOpus4_8{}, true
	case "claude-opus-4-6":
		return &ClaudeOpus4_6{}, true
	case "claude-sonnet-5":
		return &ClaudeSonnet5{}, true
	case "claude-sonnet-4-6":
		return &ClaudeSonnet4_6{}, true
	case "claude-haiku-4-5":
		return &ClaudeHaiku4_5{}, true
	}
}

// NewGenerativeModel creates a new instance of a generative model
func NewGenerativeModel(modelName, apiKey string) (assistant.GenerativeModel, error) {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	switch modelName {
	case "claude-fable-5-1":
		return NewClaudeFable5_1(client), nil
	case "claude-fable-5":
		return NewClaudeFable5(client), nil
	case "claude-opus-5-5":
		return NewClaudeOpus5_5(client), nil
	case "claude-opus-5":
		return NewClaudeOpus5(client), nil
	case "claude-opus-4-8":
		return NewClaudeOpus4_8(client), nil
	case "claude-opus-4-6":
		return NewClaudeOpus4_6(client), nil
	case "claude-sonnet-5":
		return NewClaudeSonnet5(client), nil
	case "claude-sonnet-4-6":
		return NewClaudeSonnet4_6(client), nil
	case "claude-haiku-4-5":
		return NewClaudeHaiku4_5(client), nil
	}
	return nil, fmt.Errorf("unsupported model name: %s", modelName)
}

func buildRequestBody(
	ctx context.Context,
	model anthropic.Model,
	systemInstruction []*assistant.TextContent,
	tools []assistant.ToolDefinition,
	genOpts assistant.GenerationOptions,
	msgs []assistant.Message,
) (*anthropic.MessageNewParams, error) {
	messages, err := messageParams(ctx, msgs...)
	if err != nil {
		return nil, fmt.Errorf("build message params: %w", err)
	}
	maxTokens := defaultMaxTokens
	if genOpts.MaxTokens > 0 {
		maxTokens = genOpts.MaxTokens
	}
	params := &anthropic.MessageNewParams{
		MaxTokens: anthropic.F(int64(maxTokens)),
		Model:     anthropic.F(model),
		Messages:  anthropic.F(messages),
		System:    anthropic.F(systemMessageParam(systemInstruction)),
	}
	if len(tools) > 0 {
		params.Tools = anthropic.F(toolParams(tools))
		// auto is the only tool_choice this codebase sends: forced tool use
		// ({type: "any"} / {type: "tool", name: ...}) returns 400 on
		// Claude Fable 5.1 and other current-generation models. Whether a
		// tool should be called is instead stated in the tool's own
		// Description.
		params.ToolChoice = anthropic.F[anthropic.ToolChoiceUnionParam](anthropic.ToolChoiceAutoParam{
			Type: anthropic.F(anthropic.ToolChoiceAutoTypeAuto),
		})
	}
	return params, nil
}

// requestOptions returns the per-request options that carry the parts of
// GenerationOptions the SDK's MessageNewParams has no field for.
//
// output_config.effort is injected into the serialized body because this
// SDK version predates the output_config parameter.
func requestOptions(genOpts assistant.GenerationOptions) []option.RequestOption {
	var opts []option.RequestOption
	if genOpts.Effort != "" {
		opts = append(opts, option.WithJSONSet("output_config.effort", genOpts.Effort))
	}
	return opts
}

func toolParams(defs []assistant.ToolDefinition) []anthropic.ToolUnionUnionParam {
	out := make([]anthropic.ToolUnionUnionParam, 0, len(defs))
	for _, d := range defs {
		out = append(out, anthropic.ToolParam{
			Name:        anthropic.F(d.Name),
			Description: anthropic.F(d.Description),
			InputSchema: anthropic.F[interface{}](d.InputSchema),
		})
	}
	return out
}

func messageParamFrom(ctx context.Context, src assistant.Message) (*anthropic.MessageParam, error) {
	logger := logging.LoggerFrom(ctx)

	contents := []anthropic.ContentBlockParamUnion{}
	for _, content := range src.GetContents() {
		c, err := convertContentBlockParamUnion(content)
		if err != nil {
			logger.Warn("ignore unsupported content type", "type", fmt.Sprintf("%T", content))
			continue
		}
		contents = append(contents, c)
	}
	switch src.(type) {
	case *assistant.AssistantMessage:
		m := anthropic.NewAssistantMessage(contents...)
		return &m, nil
	case *assistant.UserMessage:
		m := anthropic.NewUserMessage(contents...)
		return &m, nil
	default:
		return nil, fmt.Errorf("unknown message type: %T", src)
	}
}

func convertContentBlockParamUnion(src assistant.MessageContent) (anthropic.ContentBlockParamUnion, error) {
	switch m := src.(type) {
	case *assistant.TextContent:
		return anthropic.NewTextBlock(m.Text), nil
	case *assistant.AttachmentContent:
		return anthropic.NewTextBlock(m.ToText()), nil
	case *assistant.ToolUseContent:
		// Decode into a map rather than passing the json.RawMessage bytes
		// through directly: the SDK's interface{} field would otherwise
		// marshal a []byte as a base64 string instead of a JSON object.
		var input map[string]any
		if len(m.Input) > 0 {
			if err := json.Unmarshal(m.Input, &input); err != nil {
				return nil, fmt.Errorf("tool_use input: %w", err)
			}
		}
		return anthropic.NewToolUseBlockParam(m.ID, m.Name, input), nil
	case *assistant.ToolResultContent:
		return anthropic.NewToolResultBlock(m.ToolUseID, m.Content, m.IsError), nil
	case *assistant.ThinkingContent:
		return anthropic.ThinkingBlockParam{
			Type:      anthropic.F(anthropic.ThinkingBlockParamTypeThinking),
			Thinking:  anthropic.F(m.Thinking),
			Signature: anthropic.F(m.Signature),
		}, nil
	case *assistant.RedactedThinkingContent:
		return anthropic.RedactedThinkingBlockParam{
			Type: anthropic.F(anthropic.RedactedThinkingBlockParamTypeRedactedThinking),
			Data: anthropic.F(m.Data),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported content type: %T", src)
	}
}

func messageParams(ctx context.Context, msgs ...assistant.Message) ([]anthropic.MessageParam, error) {
	var messages []anthropic.MessageParam
	for _, msg := range msgs {
		m, err := messageParamFrom(ctx, msg)
		if err != nil {
			return nil, fmt.Errorf("anthropic message param from: %w", err)
		}
		messages = append(messages, *m)
	}
	return messages, nil
}

// toUsage converts an Anthropic Usage into the provider-agnostic assistant.Usage.
//
// Anthropic's input_tokens counts only the tokens after the last cache
// breakpoint, so the total prompt size is the sum of all three fields.
func toUsage(src anthropic.Usage) *assistant.Usage {
	return &assistant.Usage{
		InputTokens:       int(src.InputTokens + src.CacheCreationInputTokens + src.CacheReadInputTokens),
		OutputTokens:      int(src.OutputTokens),
		CachedInputTokens: int(src.CacheReadInputTokens),
		CacheWriteTokens:  int(src.CacheCreationInputTokens),
	}
}

// generateContentStream is the shared GenerateContentStream implementation
// for every Anthropic model. Each model file calls it with its own name,
// system instruction, tools and generation options.
func generateContentStream(
	ctx context.Context,
	client *anthropic.Client,
	modelName string,
	systemInstruction []*assistant.TextContent,
	tools []assistant.ToolDefinition,
	genOpts assistant.GenerationOptions,
	msgs []assistant.Message,
) (iter.Seq2[*assistant.GenerateContentResponse, error], error) {
	logger := logging.LoggerFrom(ctx).With("provider", "anthropic", "model", modelName)

	body, err := buildRequestBody(
		logging.ContextWith(ctx, logger),
		anthropic.Model(modelName),
		systemInstruction, tools, genOpts, msgs)
	if err != nil {
		return nil, fmt.Errorf("anthropic request body: %w", err)
	}
	stream := client.Messages.NewStreaming(ctx, *body, requestOptions(genOpts)...)
	return streamContent(logging.ContextWith(ctx, logger), stream), nil
}

// generateContent is the shared, non-streaming GenerateContent
// implementation. It is not exercised by the CLI today (the --no-stream
// flag is unwired), so it only covers the minimal, previously-existing
// text-only behavior: tool_use content in a non-streaming response is
// warned about and dropped rather than surfaced to the caller.
func generateContent(
	ctx context.Context,
	client *anthropic.Client,
	modelName string,
	systemInstruction []*assistant.TextContent,
	tools []assistant.ToolDefinition,
	genOpts assistant.GenerationOptions,
	msgs []assistant.Message,
) (*assistant.GenerateContentResponse, error) {
	logger := logging.LoggerFrom(ctx).With("provider", "anthropic", "model", modelName)

	body, err := buildRequestBody(
		logging.ContextWith(ctx, logger),
		anthropic.Model(modelName),
		systemInstruction, tools, genOpts, msgs)
	if err != nil {
		return nil, fmt.Errorf("anthropic request body: %w", err)
	}
	res, err := client.Messages.New(ctx, *body, requestOptions(genOpts)...)
	if err != nil {
		return nil, fmt.Errorf("anthropic New Message: %w", err)
	}

	logger = logger.With("request-id", res.ID)
	if res.StopReason != anthropic.MessageStopReasonEndTurn && res.StopReason != anthropic.MessageStopReasonToolUse {
		logger.Warn(fmt.Sprintf("anthropic response stop with reason: %s", res.StopReason))
	}
	if len(res.Content) == 0 {
		return nil, fmt.Errorf("anthropic response has no content")
	}

	var text string
	for _, c := range res.Content {
		switch c.Type {
		case anthropic.ContentBlockTypeText:
			text += c.Text
		case anthropic.ContentBlockTypeToolUse:
			logger.Warn("ignoring tool_use content in non-streaming response", "tool", c.Name)
		}
	}
	return &assistant.GenerateContentResponse{
		Content: assistant.NewTextContent(text),
		Usage:   toUsage(res.Usage),
	}, nil
}

// streamContent converts an Anthropic SSE stream into
// assistant.GenerateContentResponse items: text deltas as they arrive, one
// ToolUseContent/ThinkingContent/RedactedThinkingContent item per completed
// content block, and a terminal Usage item.
//
// The stream fails (yields an error and stops) if a tool_use block's input
// is not complete, valid JSON when the block finishes (assistant.
// ErrTruncatedToolUse — this can happen when the turn is cut off), or if the
// turn ends with stop_reason "max_tokens" (assistant.ErrMaxTokens) or
// "refusal" (assistant.ErrRefusal). In all three cases the caller must not
// treat the turn as a complete response.
func streamContent(
	ctx context.Context,
	stream *ssestream.Stream[anthropic.MessageStreamEvent],
) iter.Seq2[*assistant.GenerateContentResponse, error] {
	logger := logging.LoggerFrom(ctx)

	return func(yield func(*assistant.GenerateContentResponse, error) bool) {
		message := anthropic.Message{}
		for stream.Next() {
			event := stream.Current()
			if err := message.Accumulate(event); err != nil {
				yield(nil, fmt.Errorf("anthropic accumulate: %w", err))
				return
			}

			switch event.Type {
			case anthropic.MessageStreamEventTypeContentBlockDelta:
				if delta, ok := event.Delta.(anthropic.ContentBlockDeltaEventDelta); ok && delta.Text != "" {
					resp := &assistant.GenerateContentResponse{Content: assistant.NewTextContent(delta.Text)}
					if !yield(resp, nil) {
						return
					}
				}

			case anthropic.MessageStreamEventTypeContentBlockStop:
				idx := int(event.Index)
				if idx < 0 || idx >= len(message.Content) {
					continue
				}
				cb := message.Content[idx]

				var out assistant.MessageContent
				switch cb.Type {
				case anthropic.ContentBlockTypeToolUse:
					if !json.Valid(cb.Input) {
						yield(nil, fmt.Errorf("%w (tool %q)", assistant.ErrTruncatedToolUse, cb.Name))
						return
					}
					out = &assistant.ToolUseContent{ID: cb.ID, Name: cb.Name, Input: cb.Input}
				case anthropic.ContentBlockTypeThinking:
					out = &assistant.ThinkingContent{Thinking: cb.Thinking, Signature: cb.Signature}
				case anthropic.ContentBlockTypeRedactedThinking:
					out = &assistant.RedactedThinkingContent{Data: cb.Data}
				}
				if out != nil {
					if !yield(&assistant.GenerateContentResponse{Content: out}, nil) {
						return
					}
				}
			}
		}

		if err := stream.Err(); err != nil {
			logger.Error(fmt.Sprintf("stream error: %v", err))
			yield(nil, fmt.Errorf("anthropic stream error: %w", err))
			return
		}
		switch string(message.StopReason) {
		case "max_tokens":
			yield(nil, assistant.ErrMaxTokens)
			return
		case "refusal":
			yield(nil, assistant.ErrRefusal)
			return
		}
		yield(&assistant.GenerateContentResponse{Usage: toUsage(message.Usage)}, nil)
	}
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
