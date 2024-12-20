package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

const chatEndpoint = "https://api.openai.com/v1/chat/completions"

// chatAvailableModels is a list of models that are compatible with the Chat API
//
// https://platform.openai.com/docs/models/model-endpoint-compatibility#model-endpoint-compatibility
var chatAvailableModels = []string{
	"gpt-4o",
	"gpt-4o-mini",
	"gpt-4-turbo",
	"gpt-4",
	"gpt-3.5-turbo",

	"chatgpt-4o-latest", // continuously points to the version of GPT-4o used in ChatGPT
}

const DefaultChatModel = "chatgpt-4o-latest"

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

// Do is used to make a request to the Chat API
func (c *ChatClient) Do(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req.Stream {
		return nil, fmt.Errorf("streaming is not supported with Do, use Stream")
	}

	if req.Model == "" {
		req.Model = DefaultChatModel
	}

	resp := &ChatResponse{}
	err := c.client.doPost(ctx, chatEndpoint, req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *ChatClient) DoStream(ctx context.Context, req *ChatRequest, onReceive func(resp *ChatResponse) error) error {
	if !req.Stream {
		req.Stream = true
	}
	if req.Model == "" {
		req.Model = DefaultChatModel
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	onReceive_ := func(resp []byte) error {
		var chatResp ChatResponse
		err := json.Unmarshal(resp, &chatResp)
		if err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return onReceive(&chatResp)
	}
	return c.client.doStream(ctx, chatEndpoint, bytes.NewReader(jsonBody), onReceive_)
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
	// detect abuse. [Learn more](https://platform.openai.com/docs/guides/safety-best-practices/end-user-ids).
	User string `json:"user,omitempty"`
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
type ChatResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Usage   Usage    `json:"usage"`
	Choices []Choice `json:"choices"`
}
