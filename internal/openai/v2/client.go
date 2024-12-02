package openai

import (
	"context"
	"fmt"
	"net/http"

	"micheam.com/aico/internal/openai/gen"
)

// TODO(micheam); enable to specify the endpoint url to use OpenAI API compatible services.
const openaiEndpoint = "https://api.openai.com/v1"

// Client is used to access the OpenAI API
type Client struct {
	gen.ClientInterface
}

// NewClient returns a new Client
func NewClient(apiKey string) (*Client, error) {
	var opts []gen.ClientOption
	opts = append(opts, gen.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")
		return nil
	}))
	// TODO(micheam): Enable to log the request and response for debugging.
	// opts = append(opts, gen.WithHTTPClient( ... ))

	client, err := gen.NewClient(openaiEndpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("openai: failed to create client: %w", err)
	}
	return &Client{ClientInterface: client}, nil
}

// TODO: re-design this to be more generic
//
//	The word `Chat` is not in the API endpoint
//func (c *Client) ChatSubscribe(ctx context.Context, endpoint string, req *ChatRequest, onReceive func(resp *ChatResponse) error) error {
//	body, err := json.Marshal(req)
//	if err != nil {
//		return fmt.Errorf("failed to marshal request: %w", err)
//	}

//	httpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
//	if err != nil {
//		return fmt.Errorf("failed to create request: %w", err)
//	}

//	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
//	httpReq.Header.Set("Content-Type", "application/json")

//	httpResp, err := http.DefaultClient.Do(httpReq)
//	if err != nil {
//		return fmt.Errorf("failed to make request: %w", err)
//	}
//	defer httpResp.Body.Close()

//	if httpResp.StatusCode != http.StatusOK {
//		err := fmt.Errorf("request failed: %s", httpResp.Status)
//		// append error message from response body if any
//		b := new(bytes.Buffer)
//		b.ReadFrom(httpResp.Body)
//		if b.Len() > 0 {
//			err = fmt.Errorf("%s: %s", err, b.String())
//		}
//		return err
//	}

//	// Handle Server-Sent Events
//	scanner := bufio.NewScanner(httpResp.Body)
//	for scanner.Scan() {
//		line := scanner.Text()
//		if line == "" {
//			continue
//		}

//		// TODO: re-design this to be more generic with SSE
//		// NOTE: this depends on the sepec of the OpenAI API
//		//
//		// This handle Event Stream body as a Data-Only Event.
//		// But, real body should be a form of Event Stream Format according to the [event_stream_format]
//		//
//		// [event_stream_format]: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events#event_stream_format
//		//
//		if !strings.HasPrefix(line, "data:") {
//			return fmt.Errorf("un supported event form: %s", line)
//		}

//		chunk := strings.SplitN(line, ":", 2)
//		if len(chunk) != 2 {
//			return fmt.Errorf("un supported event form: %s", line)
//		}

//		data := strings.TrimSpace(chunk[1])
//		if data == "[DONE]" { // NOTE: this based on the sepec of the OpenAI API
//			break
//		}

//		var resp ChatResponse
//		if err := json.Unmarshal([]byte(data), &resp); err != nil {
//			return fmt.Errorf("failed to unmarshal response: %w", err)
//		}
//		if err := onReceive(&resp); err != nil {
//			return err
//		}
//	}

//	return scanner.Err()
//}
