package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"micheam.com/aico/internal/openai/gen"
	"micheam.com/aico/internal/pointer"
)

const chatEndpoint = "https://api.openai.com/v1/chat/completions"

// chatAvailableModels is a list of models that are compatible with the Chat API
//
// https://platform.openai.com/docs/models/model-endpoint-compatibility
var chatAvailableModels = []string{
	"gpt-3.5-turbo",
	"gpt-4",
	"gpt-4-turbo",
	"gpt-4o",
}

var defaultChatModel = "gpt-4o"

func ChatAvailableModels() []string {
	return chatAvailableModels
}

// ChatClient is used to access the Chat API
type ChatClient struct {
	client *Client
}

// NewChatClient returns a new ChatClient
func NewChatClient(client *Client) *ChatClient {
	return &ChatClient{client: client}
}

func (c *ChatClient) Do(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	panic("not implemented")
}

// DoStream sends a request to the Chat API and receives a stream of responses
// Specify callback function onReceive to handle each response.
//
// NOTE: Currently, multiple choices are not supported. Only one choice is supported.
// Maybe `iterators` is a better name than `onReceive` :thinking:
func (c *ChatClient) DoStream(ctx context.Context, req *ChatRequest, onReceive func(resp *ChatResponse) error) error {
	messages, err := req.ChatCompletionRequestMessage()
	if err != nil {
		return fmt.Errorf("converting messages: %w", err)
	}
	r := gen.CreateChatCompletionRequest{
		Model:    gen.CreateChatCompletionRequest_Model(req.Model),
		Messages: messages,
		Stream:   pointer.Ptr(true),
	}
	jsonReq, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}
	resp, err := c.client.CreateChatCompletionWithBody(ctx, "application/json", bytes.NewReader(jsonReq))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		body := new(bytes.Buffer)
		body.ReadFrom(resp.Body)
		return fmt.Errorf("unexpected status code: %d: %s", resp.StatusCode, body.String())
	}
	defer resp.Body.Close()

	// Handle Server-Sent Events
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var chunkResp gen.CreateChatCompletionResponse
		if err := json.Unmarshal(scanner.Bytes(), &chunkResp); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
		if len(chunkResp.Choices) == 0 {
			continue
		}
		chosen := chunkResp.Choices[0] // NOTE: Currently, only one choice is supported
		switch reason := chosen.FinishReason; reason {
		case gen.CreateChatCompletionResponseChoicesFinishReasonStop:
			return nil // Done streaming
		default:
			resp := &ChatResponse{
				ID:        chunkResp.Id,
				CreatedAt: time.Unix(int64(chunkResp.Created), 0),
				Model:     chunkResp.Model,
				Usage: Usage{
					PromptTokens:     chunkResp.Usage.PromptTokens,
					CompletionTokens: chunkResp.Usage.CompletionTokens,
					TotalTokens:      chunkResp.Usage.TotalTokens,
				},
				Delta: &DeltaMessage{
					Content: pointer.Deref(chosen.Message.Content, ""),
				},
			}
			if err := onReceive(resp); err != nil {
				return err
			}
		}
	}

	return scanner.Err()
}

// ChatRequest is used to make a request to the Chat API
//
// TODO(micheam): Abstract a bit more, because we don't need to match OpeniAI API messages exactly.
// TODO(micheam): Divide this struct into ChatRequest and ChatStreamRequest.
type ChatRequest struct {
	// model string Required
	//
	// ID of the model to use. See the model endpoint compatibility table for
	// details on which models work with the Chat API.
	Model string `json:"model"`

	// messages array Required
	//
	// A list of messages describing the conversation so far.
	Messages []Message `json:"messages"`

	// temperature number Optional Defaults to 1
	//
	// What sampling temperature to use, between 0 and 2.
	// Higher values like 0.8 will make the output more random, while lower values
	// like 0.2 will make it more focused and deterministic.
	// We generally recommend altering this or top_p but not both.
	Temperature float64 `json:"temperature,omitempty"`

	// top_p number Optional Defaults to 1
	//
	// An alternative to sampling with temperature, called nucleus sampling,
	// where the model considers the results of the tokens with top_p probability mass.
	// So 0.1 means only the tokens comprising the top 10% probability mass are considered.
	//
	// We generally recommend altering this or temperature but not both.
	TopP float64 `json:"top_p,omitempty"`

	// n integer Optional Defaults to 1
	//
	// How many chat completion choices to generate for each input message.
	N int `json:"n,omitempty"`

	// stream boolean Optional Defaults to false
	//
	// If set, partial message deltas will be sent, like in ChatGPT. Tokens will be sent as
	// data-only server-sent events as they become available, with the stream terminated by
	// a data: [DONE] message. See the OpenAI Cookbook for example code.
	Stream bool `json:"stream,omitempty"`

	// stop string or array Optional Defaults to null
	//
	// Up to 4 sequences where the API will stop generating further tokens.
	Stop []string `json:"stop,omitempty"`

	// max_tokens integer Optional Defaults to inf
	//
	// The maximum number of tokens to generate in the chat completion.
	//
	// The total length of input tokens and generated tokens is limited by the model's context length.
	MaxTokens int `json:"max_tokens,omitempty"`

	// presence_penalty number
	// Optional
	// Defaults to 0
	//
	// Number between -2.0 and 2.0. Positive values penalize new tokens based on whether they
	// appear in the text so far, increasing the model's likelihood to talk about new topics.
	//
	// See more information about frequency and presence penalties.
	PresencePenalty float64 `json:"presence_penalty,omitempty"`

	// frequency_penalty number Optional Defaults to 0
	//
	// Number between -2.0 and 2.0. Positive values penalize new tokens based on their existing
	// frequency in the text so far, decreasing the model's likelihood to repeat the same line verbatim.
	//
	// See more information about frequency and presence penalties.
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`

	// logit_bias map Optional Defaults to null
	//
	// Modify the likelihood of specified tokens appearing in the completion.
	//
	// Accepts a json object that maps tokens (specified by their token ID in the tokenizer) to
	// an associated bias value from -100 to 100. Mathematically, the bias is added to the logits
	// generated by the model prior to sampling. The exact effect will vary per model, but values
	// between -1 and 1 should decrease or increase likelihood of selection;
	// values like -100 or 100 should result in a ban or exclusive selection of the relevant token.
	LogitBias map[string]float64 `json:"logit_bias,omitempty"`

	// user string Optional
	//
	// A unique identifier representing your end-user, which can help OpenAI to monitor and
	// detect abuse. [Learn more](https://platform.openai.com/docs/guides/safety-best-practices/end-user-ids).
	User string `json:"user,omitempty"`
}

// ChatCompletionRequestMessage returns the messages in the format required by the Chat API
//
// NOTE: ChatCompletionRequestMessage is a union type that can be one of:
//   - ChatCompletionRequestUserMessage
//   - ChatCompletionRequestToolMessage
//   - ChatCompletionRequestSystemMessage
//   - ChatCompletionRequestFunctionMessage
//
// Currently, only ChatCompletionRequestUserMessage and ChatCompletionRequestSystemMessage are supported.
//
// TODO(micheam): Implement the other message types. Maybe start by learning what the Tool and Function messages are.
func (r *ChatRequest) ChatCompletionRequestMessage() ([]gen.ChatCompletionRequestMessage, error) {
	var messages []gen.ChatCompletionRequestMessage
	for i, m := range r.Messages {
		if m.Content == "" {
			return nil, fmt.Errorf("message %d has no content", i)
		}

		var msg = &gen.ChatCompletionRequestMessage{}

		switch m.Role {
		default:
			return nil, fmt.Errorf("currently message role %q is not supported, sorry...", m.Role)

		case RoleUser:
			content := &gen.ChatCompletionRequestUserMessage_Content{}
			if err := content.FromChatCompletionRequestUserMessageContent0(m.Content); err != nil {
				return nil, fmt.Errorf("user message %d to ChatCompletionRequestUserMessageContent0: %w", i, err)
			}
			// TODO(micheam): Implement the other message types. FromChatCompletionRequestUserMessageContent1

			var name *string
			if m.Name != "" {
				name = pointer.Ptr(m.Name)
			}
			value := gen.ChatCompletionRequestUserMessage{
				Content: *content,
				Name:    name,
				Role:    gen.ChatCompletionRequestUserMessageRoleUser,
			}
			if err := msg.FromChatCompletionRequestUserMessage(value); err != nil {
				return nil, fmt.Errorf("user message %d to ChatCompletionRequestMessage: %w", i, err)
			}

		case RoleSystem:
			content := &gen.ChatCompletionRequestSystemMessage_Content{}
			if err := content.FromChatCompletionRequestSystemMessageContent0(m.Content); err != nil {
				return nil, fmt.Errorf("system message %d to ChatCompletionRequestSystemMessageContent0: %w", i, err)
			}
			// TODO(micheam): Implement the other message types. FromChatCompletionRequestSystemMessageContent1

			value := gen.ChatCompletionRequestSystemMessage{
				Content: *content,
				Role:    gen.System,
			}
			if err := msg.FromChatCompletionRequestSystemMessage(value); err != nil {
				return nil, fmt.Errorf("system message %d to ChatCompletionRequestMessage: %w", i, err)
			}

		}

		messages = append(messages, *msg)
	}
	return messages, nil
}

// NewChatRequest returns a new ChatRequest.
//
// If model is empty, [defaultChatModel] will be used.
// Use [DefaultChatModel] to get the default model.
// Use [ChatAvailableModels] to get a list of available models.
func NewChatRequest(model string, messages []Message) *ChatRequest {
	return &ChatRequest{
		Model:    model,
		Messages: messages,
	}
}

// ChatResponse is the response from the Chat API
// TODO(micheam): This will divide into ChatResponse and ChatStreamResponse (or ChunkResponse)
type ChatResponse struct {
	ID        string
	CreatedAt time.Time
	Model     string
	Usage     Usage
	Choices   []Choice // Deprecated

	Index   int
	Message *Message // This will not be a part of `chat.completion.chunk` response
	Delta   *DeltaMessage
}
