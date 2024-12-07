package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strings"
)

// Client is used to access the OpenAI API
type Client struct {
	apiKey     string // APIKey string Required
	httpClient *http.Client
}

// NewClient returns a new Client
func NewClient(apiKey string) *Client {
	return &Client{
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
func (c *Client) SetHTTPClient(httpClient *http.Client) {
	c.httpClient = httpClient
}

// doPost is used to make a POST request to the OpenAI API
func (c *Client) doPost(_ context.Context, endpoint string, req any, resp any) error {
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

func (c *Client) doStream(ctx context.Context, endpoint string, body *bytes.Reader, onReceive func(resp []byte) error) error {
	httpReq, err := http.NewRequest(http.MethodPost, endpoint, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	defer httpReq.Body.Close()
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	onError := func(statusCode int, body []byte) error {
		return fmt.Errorf("got http error %s: %s", http.StatusText(statusCode), string(body))
	}
	return HandleServerSentEvents(ctx, httpReq, onReceive, onError)
}

// =================================================================================================
// Helper Functions
//

func HandleServerSentEvents(
	ctx context.Context,
	req *http.Request,
	onReceive func(resp []byte) error,
	onError func(statusCode int, body []byte) error,
) error {
	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer httpResp.Body.Close()

	// Handle HTTP Error
	if httpResp.StatusCode != http.StatusOK {
		b := new(bytes.Buffer)
		b.ReadFrom(httpResp.Body)
		return onError(httpResp.StatusCode, b.Bytes())
	}

	// Handle Server-Sent Events
	scanner := bufio.NewScanner(httpResp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			return fmt.Errorf("un supported event form: %s", line)
		}
		chunk := strings.SplitN(line, ":", 2)
		if len(chunk) != 2 {
			return fmt.Errorf("un supported event form: %s", line)
		}
		rawData := strings.TrimSpace(chunk[1])
		if rawData == "[DONE]" { // NOTE: this based on the sepec of the OpenAI API
			break
		}
		if err := onReceive([]byte(rawData)); err != nil {
			return err
		}
		// Continue to next event...
	}

	return scanner.Err()
}

// DebugTransport is a custom transport that outputs HTTP request and response
// debugging information.
type DebugTransport struct {
	Transport http.RoundTripper
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
