package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
