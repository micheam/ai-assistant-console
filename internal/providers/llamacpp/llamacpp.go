// Package llamacpp implements [complete.Completer] against the /infill endpoint
// of llama.cpp's llama-server.
package llamacpp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"micheam.com/aico/internal/complete"
)

// Options configures a Client.
type Options struct {
	// Endpoint is the full URL of the /infill endpoint.
	Endpoint string

	// Model is sent as the request's model field. Empty omits it.
	Model string

	// APIKey is sent as a bearer token. Empty sends no Authorization header.
	APIKey string

	// NPredict, TMaxPromptMS and TMaxPredictMS are sent as is.
	NPredict      int
	TMaxPromptMS  int
	TMaxPredictMS int

	// HTTPClient is used for requests. Nil means http.DefaultClient.
	HTTPClient *http.Client
}

// Client sends FIM requests to llama-server.
type Client struct {
	opts       Options
	httpClient *http.Client
}

var _ complete.Completer = (*Client)(nil)

// New returns a Client for the given options.
func New(opts Options) *Client {
	httpClient := opts.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{opts: opts, httpClient: httpClient}
}

// Sampler settings sent with every request. They are the ones llama.vim
// uses, so completions match what llama.vim would produce for the same
// context.
var samplers = []string{"top_k", "top_p", "infill"}

const (
	topK = 40
	topP = 0.90
)

// maxErrorBody caps how much of a non-200 response body is quoted in the
// returned error.
const maxErrorBody = 4 << 10

type infillChunk struct {
	Filename string `json:"filename"`
	Text     string `json:"text"`
}

type infillRequest struct {
	InputPrefix   string        `json:"input_prefix"`
	InputSuffix   string        `json:"input_suffix"`
	InputExtra    []infillChunk `json:"input_extra,omitempty"`
	Prompt        string        `json:"prompt"`
	NPredict      int           `json:"n_predict"`
	NIndent       int           `json:"n_indent,omitempty"`
	NCmpl         int           `json:"n_cmpl"`
	CachePrompt   bool          `json:"cache_prompt"`
	Stream        bool          `json:"stream"`
	TMaxPromptMS  int           `json:"t_max_prompt_ms"`
	TMaxPredictMS int           `json:"t_max_predict_ms"`
	Samplers      []string      `json:"samplers"`
	TopK          int           `json:"top_k"`
	TopP          float64       `json:"top_p"`
	Model         string        `json:"model,omitempty"`
}

type infillResponse struct {
	Content      string `json:"content"`
	Truncated    bool   `json:"truncated"`
	TokensCached int    `json:"tokens_cached"`
	Timings      struct {
		PromptN     int     `json:"prompt_n"`
		PromptMS    float64 `json:"prompt_ms"`
		PredictedN  int     `json:"predicted_n"`
		PredictedMS float64 `json:"predicted_ms"`
	} `json:"timings"`
}

// Complete sends req to /infill and returns the completion.
func (c *Client) Complete(ctx context.Context, req complete.Request) (*complete.Completion, error) {
	body, err := json.Marshal(c.buildRequest(req))
	if err != nil {
		return nil, fmt.Errorf("marshal infill request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.opts.Endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create infill request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.opts.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.opts.APIKey)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("infill request: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(httpResp.Body, maxErrorBody))
		return nil, fmt.Errorf("infill request failed: %s: %s", httpResp.Status, strings.TrimSpace(string(msg)))
	}

	resps, err := decodeResponse(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("decode infill response: %w", err)
	}
	return toCompletion(resps), nil
}

// decodeResponse reads the /infill response body: one object for a single
// completion, an array of objects when n_cmpl is greater than one.
func decodeResponse(r io.Reader) ([]infillResponse, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	trimmed := bytes.TrimLeft(data, " \t\r\n")
	if len(trimmed) > 0 && trimmed[0] == '[' {
		var resps []infillResponse
		if err := json.Unmarshal(trimmed, &resps); err != nil {
			return nil, err
		}
		if len(resps) == 0 {
			return nil, errors.New("empty completion array")
		}
		return resps, nil
	}
	var resp infillResponse
	if err := json.Unmarshal(trimmed, &resp); err != nil {
		return nil, err
	}
	return []infillResponse{resp}, nil
}

// toCompletion maps the response(s) onto a Completion. Timings and the
// cache count are taken from the first entry: all candidates share one
// prompt, so the server reports the same prompt figures for each.
func toCompletion(resps []infillResponse) *complete.Completion {
	out := &complete.Completion{
		Timings: complete.Timings{
			PromptTokens:    resps[0].Timings.PromptN,
			PromptMS:        resps[0].Timings.PromptMS,
			PredictedTokens: resps[0].Timings.PredictedN,
			PredictedMS:     resps[0].Timings.PredictedMS,
			CachedTokens:    resps[0].TokensCached,
		},
	}
	for _, r := range resps {
		out.Candidates = append(out.Candidates, complete.Candidate{Text: r.Content, Truncated: r.Truncated})
	}
	return out
}

// buildRequest maps req onto the /infill request body.
//
// /infill takes the lines before the current line (input_prefix) apart
// from the current line up to the cursor (prompt), so Prefix is split at
// its last newline.
func (c *Client) buildRequest(req complete.Request) infillRequest {
	cut := strings.LastIndexByte(req.Prefix, '\n') + 1
	var extra []infillChunk
	for _, chunk := range req.Extra {
		extra = append(extra, infillChunk{Filename: chunk.File, Text: chunk.Text})
	}
	return infillRequest{
		InputPrefix:   req.Prefix[:cut],
		InputSuffix:   req.Suffix,
		InputExtra:    extra,
		Prompt:        req.Prefix[cut:],
		NPredict:      c.opts.NPredict,
		NIndent:       req.Indent,
		NCmpl:         max(req.Count, 1),
		CachePrompt:   true,
		Stream:        false,
		TMaxPromptMS:  c.opts.TMaxPromptMS,
		TMaxPredictMS: c.opts.TMaxPredictMS,
		Samplers:      samplers,
		TopK:          topK,
		TopP:          topP,
		Model:         c.opts.Model,
	}
}
