package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Client is used to access the OpenAI API
type Client struct {
	// APIKey string Required
	APIKey string
}

// NewClient returns a new Client
func NewClient(apiKey string) *Client {
	return &Client{APIKey: apiKey}
}

// doPost is used to make a POST request to the OpenAI API
func (c *Client) doPost(ctx context.Context, endpoint string, req any, resp any) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := http.DefaultClient.Do(httpReq)
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

// TODO: re-design this to be more generic
//
//	The word `Chat` is not in the API endpoint
func (c *Client) ChatSubscribe(ctx context.Context, endpoint string, req *ChatRequest, onReceive func(resp *ChatResponse) error) error {
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := http.DefaultClient.Do(httpReq)
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

	// Handle Server-Sent Events
	scanner := bufio.NewScanner(httpResp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// TODO: re-design this to be more generic with SSE
		// NOTE: this depends on the sepec of the OpenAI API
		//
		// This handle Event Stream body as a Data-Only Event.
		// But, real body should be a form of Event Stream Format according to the [event_stream_format]
		//
		// [event_stream_format]: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events#event_stream_format
		//
		if !strings.HasPrefix(line, "data:") {
			return fmt.Errorf("un supported event form: %s", line)
		}

		chunk := strings.SplitN(line, ":", 2)
		if len(chunk) != 2 {
			return fmt.Errorf("un supported event form: %s", line)
		}

		data := strings.TrimSpace(chunk[1])
		if data == "[DONE]" { // NOTE: this based on the sepec of the OpenAI API
			break
		}

		var resp ChatResponse
		if err := json.Unmarshal([]byte(data), &resp); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
		if err := onReceive(&resp); err != nil {
			return err
		}
	}

	return scanner.Err()
}
