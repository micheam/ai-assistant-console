package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"strings"
)

// APIClient is used to access the OpenAI API
type APIClient struct {
	apiKey     string // APIKey string Required
	httpClient *http.Client
}

// NewAPIClient returns a new Client
func NewAPIClient(apiKey string) *APIClient {
	return &APIClient{
		apiKey:     apiKey,
		httpClient: http.DefaultClient,
	}
}

// SetHTTPClient is used to set the HTTP client
//
// Example: Set a custom HTTP client with a timeout of 10 seconds
//
//	client := openai.NewClient("your-api-key")
//	client.SetHTTPClient(&http.Client{Timeout: 10 * time.Second})
//
// Example: Debug HTTP requests and responses
//
//	client := openai.NewClient("your-api-key")
//	client.SetHTTPClient(&openai.DebugTransport{Transport: http.DefaultTransport})
func (c *APIClient) SetHTTPClient(httpClient *http.Client) {
	c.httpClient = httpClient
}

// DoPost is used to make a POST request to the OpenAI API
func (c *APIClient) DoPost(_ context.Context, endpoint string, req any, resp any) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	httpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		err := fmt.Errorf("request failed: %s", httpResp.Status)
		// append error message from response body if any
		b := new(bytes.Buffer)
		b.ReadFrom(httpResp.Body)
		if b.Len() > 0 {
			err = fmt.Errorf("%s: %s", err, b.String())
		}
		return err
	}

	return json.NewDecoder(httpResp.Body).Decode(resp)
}

func (c *APIClient) DoStream(ctx context.Context, endpoint string, req any) (iter.Seq[string], error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	httpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	// Handle HTTP Error
	if httpResp.StatusCode != http.StatusOK {
		b := new(bytes.Buffer)
		b.ReadFrom(httpResp.Body)
		return nil, fmt.Errorf("request failed: %s: %s", httpResp.Status, b.String())
	}
	// Handle Server-Sent Events (SSE) into iterator
	scanner := bufio.NewScanner(httpResp.Body)
	return func(yield func(string) bool) {
		defer httpResp.Body.Close()
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			chunk := strings.TrimPrefix(line, "data: ")
			if strings.TrimSpace(chunk) == "[DONE]" {
				break
			}
			if !yield(chunk) {
				break
			}
		}
	}, nil
}

// DebugTransport is a custom transport that outputs HTTP request and response
// debugging information.
type DebugTransport struct {
	Transport http.RoundTripper
}

func NewDebugTransport() *DebugTransport {
	return &DebugTransport{Transport: http.DefaultTransport}
}

var _ http.RoundTripper = (*DebugTransport)(nil)

func (d *DebugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var requestDump, responseDump []byte
	defer func() {
		fmt.Printf("Request: %s\n", requestDump)
		fmt.Printf("Response: %s\n", responseDump)
	}()
	requestDump, _ = httputil.DumpRequest(req, true)
	resp, err := d.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	responseDump, _ = httputil.DumpResponse(resp, true)
	return resp, nil
}

// Message is a message in the conversation.
//
// Message can be one of the following:
//   - System message
//   - User message
//   - Assistant message
type Message interface {
	Type() string
	json.Marshaler
	json.Unmarshaler
}

type SystemMessage struct {
	Content string
	Name    *string
}

type UserMessage struct {
	Content []Content
	Name    *string
}

type AssistantMessage struct {
	Content []Content
	Name    *string
}

var _ Message = (*SystemMessage)(nil)
var _ Message = (*UserMessage)(nil)
var _ Message = (*AssistantMessage)(nil)

func (m UserMessage) Type() string      { return "user" }
func (m SystemMessage) Type() string    { return "system" }
func (m AssistantMessage) Type() string { return "assistant" }

func (msg SystemMessage) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"role":    RoleSystem,
		"content": msg.Content,
	}
	if n := msg.Name; n != nil {
		m["name"] = *n
	}
	return json.Marshal(m)
}

func (msg UserMessage) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"role":    RoleUser,
		"content": msg.Content,
	}
	if n := msg.Name; n != nil {
		m["name"] = *n
	}
	return json.Marshal(m)
}

func (msg AssistantMessage) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"role":    RoleAssistant,
		"content": msg.Content,
	}
	if n := msg.Name; n != nil {
		m["name"] = *n
	}
	return json.Marshal(m)
}

func (msg *SystemMessage) UnmarshalJSON(b []byte) error {
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	msg.Content = m["content"].(string)
	if n, ok := m["name"].(string); ok {
		msg.Name = &n
	}
	return nil
}

func (msg *UserMessage) UnmarshalJSON(b []byte) error {
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	// NOTE: 'content' will be an array of strings or single string.
	//       We need to handle both cases.
	v := reflect.ValueOf(m["content"])
	switch v.Kind() {
	case reflect.String:
		msg.Content = []Content{&TextContent{Text: v.String()}}
	case reflect.Slice:
		msg.Content = make([]Content, v.Len())
		for i := 0; i < v.Len(); i++ {
			msg.Content = append(msg.Content, &TextContent{Text: v.Index(i).String()})
		}
	default:
		return fmt.Errorf("unexpected type for 'content': %T", m["content"])
	}
	if n, ok := m["name"].(string); ok {
		msg.Name = &n
	}
	return nil
}

func (msg *AssistantMessage) UnmarshalJSON(b []byte) error {
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	// NOTE: 'content' will be an array of strings or single string.
	//       We need to handle both cases.
	v := reflect.ValueOf(m["content"])
	switch v.Kind() {
	case reflect.String:
		msg.Content = []Content{&TextContent{Text: v.String()}}
	case reflect.Slice:
		msg.Content = make([]Content, v.Len())
		for i := 0; i < v.Len(); i++ {
			msg.Content = append(msg.Content, &TextContent{Text: v.Index(i).String()})
		}
	default:
		return fmt.Errorf("unexpected type for 'content': %T", m["content"])
	}
	if n, ok := m["name"].(string); ok {
		msg.Name = &n
	}
	return nil
}

type DeltaMessage struct {
	// content string Required
	//
	// The contents of the message.
	Content string `json:"content"`
}

// The role of the author of this message.
// One of system, user, or assistant.
type Role int

const (
	RoleUndefined Role = iota
	RoleSystem
	RoleUser
	RoleAssistant
)

var _ json.Marshaler = RoleSystem
var _ json.Unmarshaler = (*Role)(nil)

func (r Role) String() string {
	return [...]string{"undefined", "system", "user", "assistant"}[r]
}

func ParseRole(s string) Role {
	switch s {
	default:
		return RoleUndefined
	case "system":
		return RoleSystem
	case "user":
		return RoleUser
	case "assistant":
		return RoleAssistant
	}
}

func (r Role) MarshalJSON() ([]byte, error) {
	return []byte(`"` + r.String() + `"`), nil
}

func (r *Role) UnmarshalJSON(b []byte) error {
	switch string(b) {
	default:
		*r = RoleUndefined
	case `"system"`:
		*r = RoleSystem
	case `"user"`:
		*r = RoleUser
	case `"assistant"`:
		*r = RoleAssistant
	}
	return nil
}

// Usage is the usage of the Chat API
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Choice is the choice of the Chat API
type Choice struct {
	Message      AssistantMessage `json:"message,omitempty"`
	Delta        *DeltaMessage    `json:"delta,omitempty"`
	FinishReason string           `json:"finish_reason"`
	Index        int              `json:"index"`
}

// ------------------------------------------------------------------------------------------------
// Content Value of the message
// ------------------------------------------------------------------------------------------------

type Content interface {
	Type() string
	json.Marshaler
	json.Unmarshaler
}

// Variations of the content:

type TextContent struct{ Text string }
type ImageContent struct{ URL url.URL }
type AudioContent struct {
	Data   []byte // base64 encoded audio data
	Format string // wav, mp3
}
type RefusalContent struct{ Refusal string }

var _ Content = (*TextContent)(nil)
var _ Content = (*ImageContent)(nil)
var _ Content = (*AudioContent)(nil)
var _ Content = (*RefusalContent)(nil)

func (c TextContent) Type() string    { return "text" }
func (c ImageContent) Type() string   { return "image_url" }
func (c AudioContent) Type() string   { return "audio" }
func (c RefusalContent) Type() string { return "refusal" }

func (c TextContent) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"type": c.Type(),
		"text": c.Text,
	})
}

// Example:
//
//	{
//	  "type": "image_url",
//	  "image_url": {
//	    "url": "<url-or-base64-encoded-image-content>",
//	    "details": "<optional-details>"
//	  },
//	}
func (c ImageContent) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type":      c.Type(),
		"image_url": map[string]string{"url": c.URL.String()},
	})
}

func (c AudioContent) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type": c.Type(),
		"input_audio": map[string]string{
			"data":   string(c.Data),
			"format": c.Format,
		},
	})
}

func (c RefusalContent) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"type":    c.Type(),
		"content": c.Refusal,
	})
}

func (c *TextContent) UnmarshalJSON(b []byte) error {
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	c.Text = m["text"]
	return nil
}

func (c *ImageContent) UnmarshalJSON(b []byte) error {
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	u, err := url.Parse(m["content"])
	if err != nil {
		return err
	}
	c.URL = *u
	return nil
}

func (c *AudioContent) UnmarshalJSON(b []byte) error {
	var m map[string]map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	c.Data = []byte(m["input_audio"]["data"])
	c.Format = m["input_audio"]["format"]
	return nil
}

func (c *RefusalContent) UnmarshalJSON(b []byte) error {
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	c.Refusal = m["content"]
	return nil
}
