package llamacpp

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"micheam.com/aico/internal/complete"
)

// capturedRequest is what the fake /infill server saw.
type capturedRequest struct {
	method string
	path   string
	header http.Header
	body   string
}

// newFakeServer starts a fake llama-server that records each request and
// answers with the given status and body.
func newFakeServer(t *testing.T, status int, respBody string) (*httptest.Server, *capturedRequest) {
	t.Helper()
	got := &capturedRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		*got = capturedRequest{method: r.Method, path: r.URL.Path, header: r.Header.Clone(), body: string(b)}
		w.WriteHeader(status)
		_, _ = io.WriteString(w, respBody)
	}))
	t.Cleanup(srv.Close)
	return srv, got
}

const okResponse = `{
	"content": "a + b",
	"truncated": true,
	"tokens_cached": 210,
	"timings": {"prompt_n": 12, "prompt_ms": 35.5, "predicted_n": 4, "predicted_ms": 48.25, "predicted_per_second": 82.9}
}`

func TestClient_Complete_RequestBody(t *testing.T) {
	tests := []struct {
		name     string
		opts     Options
		req      complete.Request
		wantBody string
	}{
		{
			name: "splits prefix at its last newline and sends every field",
			opts: Options{Model: "qwen2.5-coder-7b", NPredict: 128, TMaxPromptMS: 500, TMaxPredictMS: 1000},
			req: complete.Request{
				Prefix: "func add(a, b int) int {\n\treturn ",
				Suffix: "\n}\n",
				Indent: 1,
				Extra:  []complete.Chunk{{File: "util.go", Text: "package main\n"}},
				Count:  3,
			},
			wantBody: `{
				"input_prefix": "func add(a, b int) int {\n",
				"input_suffix": "\n}\n",
				"input_extra": [{"filename": "util.go", "text": "package main\n"}],
				"prompt": "\treturn ",
				"n_predict": 128,
				"n_indent": 1,
				"n_cmpl": 3,
				"cache_prompt": true,
				"stream": false,
				"t_max_prompt_ms": 500,
				"t_max_predict_ms": 1000,
				"samplers": ["top_k", "top_p", "infill"],
				"top_k": 40,
				"top_p": 0.9,
				"model": "qwen2.5-coder-7b"
			}`,
		},
		{
			name: "prefix without newline goes entirely to prompt; optional fields omitted",
			opts: Options{NPredict: 16, TMaxPromptMS: 1, TMaxPredictMS: 2},
			req:  complete.Request{Prefix: "fmt.", Suffix: ""},
			wantBody: `{
				"input_prefix": "",
				"input_suffix": "",
				"prompt": "fmt.",
				"n_predict": 16,
				"n_cmpl": 1,
				"cache_prompt": true,
				"stream": false,
				"t_max_prompt_ms": 1,
				"t_max_predict_ms": 2,
				"samplers": ["top_k", "top_p", "infill"],
				"top_k": 40,
				"top_p": 0.9
			}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, got := newFakeServer(t, http.StatusOK, okResponse)
			tt.opts.Endpoint = srv.URL + "/infill"

			_, err := New(tt.opts).Complete(context.Background(), tt.req)
			require.NoError(t, err)
			assert.Equal(t, http.MethodPost, got.method)
			assert.Equal(t, "/infill", got.path)
			assert.Equal(t, "application/json", got.header.Get("Content-Type"))
			assert.JSONEq(t, tt.wantBody, got.body)
		})
	}
}

func TestClient_Complete_Authorization(t *testing.T) {
	tests := []struct {
		name   string
		apiKey string
		want   string
	}{
		{name: "with API key", apiKey: "secret", want: "Bearer secret"},
		{name: "without API key", apiKey: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, got := newFakeServer(t, http.StatusOK, okResponse)
			c := New(Options{Endpoint: srv.URL + "/infill", APIKey: tt.apiKey})

			_, err := c.Complete(context.Background(), complete.Request{Prefix: "x"})
			require.NoError(t, err)
			assert.Equal(t, tt.want, got.header.Get("Authorization"))
		})
	}
}

func TestClient_Complete_Response(t *testing.T) {
	wantTimings := complete.Timings{PromptTokens: 12, PromptMS: 35.5, PredictedTokens: 4, PredictedMS: 48.25, CachedTokens: 210}

	t.Run("single object", func(t *testing.T) {
		srv, _ := newFakeServer(t, http.StatusOK, okResponse)

		got, err := New(Options{Endpoint: srv.URL + "/infill"}).Complete(context.Background(), complete.Request{Prefix: "x"})
		require.NoError(t, err)
		assert.Equal(t, &complete.Completion{
			Candidates: []complete.Candidate{{Text: "a + b", Truncated: true}},
			Timings:    wantTimings,
		}, got)
	})

	t.Run("array for several candidates", func(t *testing.T) {
		srv, _ := newFakeServer(t, http.StatusOK, "\n ["+okResponse+`, {"content": "a - b", "truncated": false, "tokens_cached": 210, "timings": {"prompt_n": 12, "prompt_ms": 35.5, "predicted_n": 4, "predicted_ms": 48.25}}]`)

		got, err := New(Options{Endpoint: srv.URL + "/infill"}).Complete(context.Background(), complete.Request{Prefix: "x", Count: 2})
		require.NoError(t, err)
		assert.Equal(t, &complete.Completion{
			Candidates: []complete.Candidate{{Text: "a + b", Truncated: true}, {Text: "a - b"}},
			Timings:    wantTimings,
		}, got)
	})

	t.Run("empty array", func(t *testing.T) {
		srv, _ := newFakeServer(t, http.StatusOK, "[]")

		_, err := New(Options{Endpoint: srv.URL + "/infill"}).Complete(context.Background(), complete.Request{Prefix: "x", Count: 2})
		require.ErrorContains(t, err, "empty completion array")
	})
}

func TestClient_Complete_Errors(t *testing.T) {
	t.Run("non-200 status quotes the status and body", func(t *testing.T) {
		srv, _ := newFakeServer(t, http.StatusServiceUnavailable, `{"error": {"message": "Loading model"}}`)

		_, err := New(Options{Endpoint: srv.URL + "/infill"}).Complete(context.Background(), complete.Request{Prefix: "x"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "503")
		assert.Contains(t, err.Error(), "Loading model")
	})

	t.Run("malformed response body", func(t *testing.T) {
		srv, _ := newFakeServer(t, http.StatusOK, `not json`)

		_, err := New(Options{Endpoint: srv.URL + "/infill"}).Complete(context.Background(), complete.Request{Prefix: "x"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "decode infill response")
	})

	t.Run("cancelled context stops the request", func(t *testing.T) {
		arrived := make(chan struct{})
		release := make(chan struct{})
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.ReadAll(r.Body)
			close(arrived)
			select {
			case <-r.Context().Done():
			case <-release:
			}
		}))
		t.Cleanup(srv.Close)
		t.Cleanup(func() { close(release) }) // runs before srv.Close

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			<-arrived
			cancel()
		}()
		_, err := New(Options{Endpoint: srv.URL + "/infill"}).Complete(ctx, complete.Request{Prefix: "x"})
		require.ErrorIs(t, err, context.Canceled)
	})
}
