package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"micheam.com/aico/internal/logging"
	"micheam.com/aico/internal/openai"
)

const endpoint = "https://api.openai.com/v1/chat/completions"

// Client is used to access the Chat API
type Client struct {
	client *openai.Client
}

// NewWithOpenAIClient returns a new ChatClient
func NewWithOpenAIClient(client *openai.Client) *Client {
	return &Client{client: client}
}

// New returns a new ChatClient.
//
// If you have an openai.Client already, use [NewWithOpenAIClient].
func New(apikey string) *Client {
	openapiClient := openai.NewClient(apikey)
	return NewWithOpenAIClient(openapiClient)
}

// Do is used to make a request to the Chat API
func (c *Client) Do(ctx context.Context, req *Request) (*Response, error) {
	logger := logging.LoggerFrom(ctx)
	if req.Stream {
		return nil, fmt.Errorf("streaming is not supported with Do, use Stream")
	}
	if req.Model == ModelUnspecified {
		logger.Warn(fmt.Sprintf("Model is not specified, using default model: %s", DefaultModel))
		req.Model = DefaultModel
	}

	resp := &Response{}
	err := c.client.DoPost(ctx, endpoint, req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) DoStream(ctx context.Context, req *Request, onReceive func(resp *Response) error) error {
	logger := logging.LoggerFrom(ctx)
	if !req.Stream {
		req.Stream = true
	}
	if req.Model == ModelUnspecified {
		logger.Warn(fmt.Sprintf("Model is not specified, using default model: %s", DefaultModel))
		req.Model = DefaultModel
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	onReceive_ := func(b []byte) error {
		var resp Response
		err := json.Unmarshal(b, &resp)
		if err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return onReceive(&resp)
	}
	return c.client.DoStream(ctx, endpoint, bytes.NewReader(jsonBody), onReceive_)
}

// Request is used to make a request to the Chat API
type Request struct {
	// model string Required
	//
	// ID of the model to use. See the model endpoint compatibility table for
	// details on which models work with the Chat API.
	Model Model `json:"model"`

	// messages array Required
	//
	// A list of messages describing the conversation so far.
	Messages []openai.Message `json:"messages"`

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

// NewRequest returns a new Request.
func NewRequest(msgs []openai.Message, opts ...func(*Request)) *Request {
	req := &Request{Messages: msgs, Model: DefaultModel}
	for _, opt := range opts {
		opt(req)
	}
	return req
}

func WithModel(model Model) func(*Request) {
	return func(req *Request) {
		req.Model = model
	}
}

// Response is the response from the Chat API
type Response struct {
	ID      string          `json:"id"`
	Object  string          `json:"object"`
	Created int64           `json:"created"`
	Model   string          `json:"model"`
	Usage   openai.Usage    `json:"usage"`
	Choices []openai.Choice `json:"choices"`
}
