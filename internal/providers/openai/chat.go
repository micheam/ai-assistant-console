package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/logging"
)

const endpoint = "https://api.openai.com/v1/chat/completions"
const ProviderName = "openai"

// AvailableModels returns a list of available models
func AvailableModels() []assistant.ModelDescriptor {
	return []assistant.ModelDescriptor{
		&GPT52{},
		&GPT41{},
		&GPT41Mini{},
		&O3{},
		&O4Mini{},
		&O3Mini{},
		// Deprecated
		&GPT4O{DeprecationInfo: assistant.DeprecationInfo{IsDeprecated: true, RemovedIn: "v0.2"}},
		&GPT4OMini{DeprecationInfo: assistant.DeprecationInfo{IsDeprecated: true, RemovedIn: "v0.2"}},
		&O1{DeprecationInfo: assistant.DeprecationInfo{IsDeprecated: true, RemovedIn: "v0.2"}},
		&O1Mini{DeprecationInfo: assistant.DeprecationInfo{IsDeprecated: true, RemovedIn: "v0.2"}},
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
	case "gpt-5.2":
		return &GPT52{}, true
	case "gpt-4.1":
		return &GPT41{}, true
	case "gpt-4.1-mini":
		return &GPT41Mini{}, true
	case "o3":
		return &O3{}, true
	case "o4-mini":
		return &O4Mini{}, true
	case "o3-mini":
		return &O3Mini{}, true
	// Deprecated
	case "gpt-4o":
		return &GPT4O{}, true
	case "gpt-4o-mini":
		return &GPT4OMini{}, true
	case "o1":
		return &O1{}, true
	case "o1-mini":
		return &O1Mini{}, true
	}
}

// NewGenerativeModel creates a new instance of a generative model
func NewGenerativeModel(modelName, apiKey string) (assistant.GenerativeModel, error) {
	switch modelName {
	case "gpt-5.2":
		return NewGPT52(apiKey), nil
	case "gpt-4.1":
		return NewGPT41(apiKey), nil
	case "gpt-4.1-mini":
		return NewGPT41Mini(apiKey), nil
	case "o3":
		return NewO3(apiKey), nil
	case "o4-mini":
		return NewO4Mini(apiKey), nil
	case "o3-mini":
		return NewO3Mini(apiKey), nil
	// Deprecated
	case "gpt-4o":
		return NewGPT4O(apiKey), nil
	case "gpt-4o-mini":
		return NewGPT4OMini(apiKey), nil
	case "o1":
		return NewO1(apiKey), nil
	case "o1-mini":
		return NewO1Mini(apiKey), nil
	}
	return nil, fmt.Errorf("unsupported model name: %s", modelName)
}

// ChatRequest is used to make a request to the Chat API
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
	// detect abuse. [Learn more](https://platform.com/docs/guides/safety-best-practices/end-user-ids).
	User string `json:"user,omitempty"`
}

// ChatResponse is the response from the Chat API
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Usage   Usage    `json:"usage"`
	Choices []Choice `json:"choices"`
}

// BuildChatRequest builds a chat request for OpenAI-compatible APIs
func BuildChatRequest(ctx context.Context, modelName string, systemInstruction []*assistant.TextContent, msgs []*assistant.Message) (*ChatRequest, error) {
	if len(msgs) == 0 {
		return nil, fmt.Errorf("no messages provided")
	}
	req := &ChatRequest{
		Model:    modelName,
		Messages: make([]Message, 0, len(msgs)+1),
	}
	if len(systemInstruction) > 0 {
		for _, c := range systemInstruction {
			req.Messages = append(req.Messages, &SystemMessage{Content: c.Text})
		}
	}
	for _, msg := range msgs {
		for _, content := range msg.Contents {
			switch v := content.(type) {
			case *assistant.AttachmentContent:
				content := []Content{&TextContent{
					Text: v.ToText(),
				}}
				switch msg.Author {
				case "user":
					req.Messages = append(req.Messages, &UserMessage{Content: content})
				case "assistant":
					req.Messages = append(req.Messages, &AssistantMessage{Content: content})
				}

			case *assistant.TextContent:
				switch msg.Author {
				case "user":
					req.Messages = append(req.Messages, &UserMessage{
						Content: []Content{&TextContent{Text: v.Text}},
					})
				case "assistant":
					req.Messages = append(req.Messages, &AssistantMessage{
						Content: []Content{&TextContent{Text: v.Text}},
					})
				}

			case *assistant.URLImageContent:
				switch msg.Author {
				case "user":
					req.Messages = append(req.Messages, &UserMessage{
						Content: []Content{&ImageContent{URL: v.URL}},
					})
				case "assistant":
					req.Messages = append(req.Messages, &AssistantMessage{
						Content: []Content{&ImageContent{URL: v.URL}},
					})
				}

			default:
				// fmt.Printf("Unsupported message content type: %s\n", reflect.TypeOf(v))
				logging.LoggerFrom(ctx).Warn(fmt.Sprintf("Unsupported message content type: %T", v))
			}
		}
	}
	return req, nil
}

// ToGenerateContentResponse converts an OpenAI ChatResponse to a GenerateContentResponse
func ToGenerateContentResponse(src *ChatResponse) *assistant.GenerateContentResponse {
	if len(src.Choices) > 0 {
		text := src.Choices[0].Message.Content[0].(*TextContent).Text
		return &assistant.GenerateContentResponse{Content: &assistant.TextContent{Text: text}}
	}
	return &assistant.GenerateContentResponse{Content: nil}
}

// GenerateContent is a shared implementation for generating content with OpenAI-compatible APIs
func GenerateContent(ctx context.Context, client *APIClient, apiEndpoint string, modelName string, systemInstruction []*assistant.TextContent, msgs []*assistant.Message) (*assistant.GenerateContentResponse, error) {
	req, err := BuildChatRequest(ctx, modelName, systemInstruction, msgs)
	if err != nil {
		return nil, fmt.Errorf("build chat request: %w", err)
	}
	resp := new(ChatResponse)
	if err := client.DoPost(ctx, apiEndpoint, req, resp); err != nil {
		return nil, err
	}
	return ToGenerateContentResponse(resp), nil
}

// GenerateContentStream is a shared implementation for streaming content with OpenAI-compatible APIs
func GenerateContentStream(ctx context.Context, client *APIClient, apiEndpoint string, modelName string, systemInstruction []*assistant.TextContent, msgs []*assistant.Message) (iter.Seq[*assistant.GenerateContentResponse], error) {
	req, err := BuildChatRequest(ctx, modelName, systemInstruction, msgs)
	if err != nil {
		return nil, fmt.Errorf("build chat request: %w", err)
	}
	req.Stream = true
	iter, err := client.DoStream(ctx, apiEndpoint, req)
	if err != nil {
		return nil, err
	}
	return func(yield func(*assistant.GenerateContentResponse) bool) {
		for s := range iter {
			var res *ChatResponse
			err := json.Unmarshal([]byte(s), &res)
			if err != nil {
				logging.LoggerFrom(ctx).Error(fmt.Sprintf("error: %v", err))
				continue
			}
			if len(res.Choices) == 0 || res.Choices[0].Delta == nil {
				continue
			}
			delta := assistant.NewTextContent(res.Choices[0].Delta.Content)
			if !yield(&assistant.GenerateContentResponse{Content: delta}) {
				break
			}
		}
	}, nil
}
