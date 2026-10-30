package serve

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"micheam.com/aico/internal/complete"
	"micheam.com/aico/internal/config"
)

// fakeCompleter answers after release is closed (or immediately when
// release is nil), or with err, and records each request's context.
type fakeCompleter struct {
	release chan struct{}
	err     error

	mu    sync.Mutex
	calls []complete.Request
}

func (f *fakeCompleter) Complete(ctx context.Context, req complete.Request) (*complete.Completion, error) {
	f.mu.Lock()
	f.calls = append(f.calls, req)
	f.mu.Unlock()
	if f.release != nil {
		select {
		case <-f.release:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if f.err != nil {
		return nil, f.err
	}
	var cands []complete.Candidate
	for i := 0; i < req.Count; i++ {
		cands = append(cands, complete.Candidate{Text: fmt.Sprintf("done:%s:%d", req.Prefix, i)})
	}
	return &complete.Completion{Candidates: cands}, nil
}

// client is a minimal JSON-RPC client over one end of a pipe.
type client struct {
	t    *testing.T
	conn net.Conn
	r    *bufio.Reader
	resp chan response
}

func testConfig() *config.Config {
	return &config.Config{Complete: config.CompleteConfig{
		DefaultProfile: "local",
		Profiles: map[string]config.CompleteProfile{
			"local": {Endpoint: "http://127.0.0.1:8012/infill", Model: "m-local", Candidates: 2},
			"other": {Endpoint: "http://rack:8012/infill"},
		},
	}}
}

func newTestServer(t *testing.T, completer complete.Completer) *Server {
	t.Helper()
	s := New(Options{
		Config:       testConfig(),
		Version:      "test",
		NewCompleter: func(config.CompleteProfile) complete.Completer { return completer },
	})
	t.Cleanup(s.Shutdown)
	return s
}

func connect(t *testing.T, s *Server) *client {
	t.Helper()
	serverSide, clientSide := net.Pipe()
	go s.ServeConn(serverSide)
	c := &client{t: t, conn: clientSide, r: bufio.NewReader(clientSide), resp: make(chan response, 16)}
	go func() {
		for {
			body, err := readMessage(c.r)
			if err != nil {
				close(c.resp)
				return
			}
			var r response
			if err := json.Unmarshal(body, &r); err != nil {
				c.t.Errorf("bad response %s: %v", body, err)
				return
			}
			c.resp <- r
		}
	}()
	t.Cleanup(func() { clientSide.Close() })
	return c
}

func (c *client) send(msg string) {
	c.t.Helper()
	require.NoError(c.t, writeMessage(c.conn, []byte(msg)))
}

func (c *client) recv() response {
	c.t.Helper()
	select {
	case r, ok := <-c.resp:
		require.True(c.t, ok, "connection closed")
		return r
	case <-time.After(5 * time.Second):
		c.t.Fatal("timed out waiting for a response")
		return response{}
	}
}

func (c *client) initialize(profile string) response {
	c.t.Helper()
	c.send(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocol_version":1,"client":{"name":"t","version":"0"},"profile":"` + profile + `"}}`)
	return c.recv()
}

func resultJSON(t *testing.T, r response) string {
	t.Helper()
	require.Nil(t, r.Error, "unexpected error: %+v", r.Error)
	b, err := json.Marshal(r.Result)
	require.NoError(t, err)
	return string(b)
}

func TestInitialize(t *testing.T) {
	t.Run("returns server info and the profile", func(t *testing.T) {
		c := connect(t, newTestServer(t, &fakeCompleter{}))
		r := c.initialize("local")
		assert.Equal(t, json.RawMessage("1"), r.ID)
		assert.JSONEq(t, `{
			"protocol_version": 1,
			"server": {"name": "aico", "version": "test"},
			"methods": ["complete"],
			"profile": {"name": "local", "model": "m-local", "endpoint": "http://127.0.0.1:8012/infill", "prefix_lines": 256, "suffix_lines": 64, "candidates": 2}
		}`, resultJSON(t, r))
	})

	t.Run("empty profile selects the default", func(t *testing.T) {
		c := connect(t, newTestServer(t, &fakeCompleter{}))
		r := c.initialize("")
		assert.Contains(t, resultJSON(t, r), `"name":"local"`)
	})

	t.Run("version mismatch", func(t *testing.T) {
		c := connect(t, newTestServer(t, &fakeCompleter{}))
		c.send(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocol_version":2}}`)
		r := c.recv()
		require.NotNil(t, r.Error)
		assert.Equal(t, codeVersionMismatch, r.Error.Code)
		assert.Equal(t, map[string]any{"protocol_version": float64(1)}, r.Error.Data)
	})

	t.Run("unknown profile", func(t *testing.T) {
		c := connect(t, newTestServer(t, &fakeCompleter{}))
		r := c.initialize("nope")
		require.NotNil(t, r.Error)
		assert.Equal(t, codeUnknownProfile, r.Error.Code)
	})

	t.Run("twice on one connection", func(t *testing.T) {
		c := connect(t, newTestServer(t, &fakeCompleter{}))
		require.Nil(t, c.initialize("local").Error)
		r := c.initialize("local")
		require.NotNil(t, r.Error)
		assert.Equal(t, codeInvalidRequest, r.Error.Code)
	})
}

func TestDispatch(t *testing.T) {
	t.Run("complete before initialize", func(t *testing.T) {
		c := connect(t, newTestServer(t, &fakeCompleter{}))
		c.send(`{"jsonrpc":"2.0","id":7,"method":"complete","params":{"prefix":"x"}}`)
		r := c.recv()
		assert.Equal(t, json.RawMessage("7"), r.ID)
		require.NotNil(t, r.Error)
		assert.Equal(t, codeNotInitialized, r.Error.Code)
	})

	t.Run("unknown method", func(t *testing.T) {
		c := connect(t, newTestServer(t, &fakeCompleter{}))
		c.send(`{"jsonrpc":"2.0","id":"abc","method":"nope"}`)
		r := c.recv()
		assert.Equal(t, json.RawMessage(`"abc"`), r.ID)
		require.NotNil(t, r.Error)
		assert.Equal(t, codeMethodNotFound, r.Error.Code)
	})

	t.Run("parse error answers with null id", func(t *testing.T) {
		c := connect(t, newTestServer(t, &fakeCompleter{}))
		c.send(`{not json`)
		r := c.recv()
		assert.Equal(t, json.RawMessage("null"), r.ID)
		require.NotNil(t, r.Error)
		assert.Equal(t, codeParseError, r.Error.Code)
	})

	t.Run("unknown notification is ignored", func(t *testing.T) {
		c := connect(t, newTestServer(t, &fakeCompleter{}))
		c.send(`{"jsonrpc":"2.0","method":"whatever"}`)
		require.Nil(t, c.initialize("local").Error)
	})
}

func TestShutdown(t *testing.T) {
	t.Run("without initialize, as the version-mismatch recovery path needs", func(t *testing.T) {
		s := newTestServer(t, &fakeCompleter{})
		c := connect(t, s)
		c.send(`{"jsonrpc":"2.0","id":1,"method":"shutdown"}`)
		require.Nil(t, c.recv().Error)
		select {
		case <-s.Done():
		case <-time.After(5 * time.Second):
			t.Fatal("server did not shut down")
		}
	})

	s := newTestServer(t, &fakeCompleter{})
	c1 := connect(t, s)
	c2 := connect(t, s)
	require.Nil(t, c1.initialize("local").Error)

	c1.send(`{"jsonrpc":"2.0","id":2,"method":"shutdown"}`)
	r := c1.recv()
	require.Nil(t, r.Error)

	select {
	case <-s.Done():
	case <-time.After(5 * time.Second):
		t.Fatal("server did not shut down")
	}
	// Both connections are closed by the server.
	for _, c := range []*client{c1, c2} {
		select {
		case _, ok := <-c.resp:
			assert.False(t, ok)
		case <-time.After(5 * time.Second):
			t.Fatal("connection not closed")
		}
	}
}

func TestComplete(t *testing.T) {
	t.Run("returns the profile's number of candidates by default", func(t *testing.T) {
		c := connect(t, newTestServer(t, &fakeCompleter{}))
		require.Nil(t, c.initialize("local").Error)
		c.send(`{"jsonrpc":"2.0","id":2,"method":"complete","params":{"prefix":"a","suffix":"b","indent":1,"extra":[{"file":"f","text":"t"}]}}`)
		r := c.recv()
		assert.JSONEq(t, `{"candidates":[{"text":"done:a:0","truncated":false},{"text":"done:a:1","truncated":false}],"timings":{"prompt_tokens":0,"prompt_ms":0,"predicted_tokens":0,"predicted_ms":0,"cached_tokens":0}}`, resultJSON(t, r))
	})

	t.Run("count overrides the profile", func(t *testing.T) {
		c := connect(t, newTestServer(t, &fakeCompleter{}))
		require.Nil(t, c.initialize("local").Error)
		c.send(`{"jsonrpc":"2.0","id":2,"method":"complete","params":{"prefix":"a","count":1}}`)
		assert.Contains(t, resultJSON(t, c.recv()), `"candidates":[{"text":"done:a:0","truncated":false}]`)

		c.send(`{"jsonrpc":"2.0","id":3,"method":"complete","params":{"prefix":"a","count":-1}}`)
		r := c.recv()
		require.NotNil(t, r.Error)
		assert.Equal(t, codeInvalidParams, r.Error.Code)
	})

	t.Run("backend error", func(t *testing.T) {
		c := connect(t, newTestServer(t, &fakeCompleter{err: errors.New("boom")}))
		require.Nil(t, c.initialize("local").Error)
		c.send(`{"jsonrpc":"2.0","id":2,"method":"complete","params":{"prefix":"a"}}`)
		r := c.recv()
		require.NotNil(t, r.Error)
		assert.Equal(t, codeBackendError, r.Error.Code)
		assert.Contains(t, r.Error.Message, "boom")
	})

	t.Run("invalid params", func(t *testing.T) {
		c := connect(t, newTestServer(t, &fakeCompleter{}))
		require.Nil(t, c.initialize("local").Error)
		c.send(`{"jsonrpc":"2.0","id":2,"method":"complete","params":{"prefix":1}}`)
		r := c.recv()
		require.NotNil(t, r.Error)
		assert.Equal(t, codeInvalidParams, r.Error.Code)
	})

	t.Run("$/cancelRequest cancels by params.id", func(t *testing.T) {
		f := &fakeCompleter{release: make(chan struct{})}
		c := connect(t, newTestServer(t, f))
		require.Nil(t, c.initialize("local").Error)
		c.send(`{"jsonrpc":"2.0","id":2,"method":"complete","params":{"prefix":"a"}}`)
		waitCalls(t, f, 1)
		c.send(`{"jsonrpc":"2.0","method":"$/cancelRequest","params":{"id":2}}`)
		r := c.recv()
		assert.Equal(t, json.RawMessage("2"), r.ID)
		require.NotNil(t, r.Error)
		assert.Equal(t, codeCancelled, r.Error.Code)
	})

	t.Run("newer request on another connection cancels the older one", func(t *testing.T) {
		f := &fakeCompleter{release: make(chan struct{})}
		s := newTestServer(t, f)
		c1, c2 := connect(t, s), connect(t, s)
		require.Nil(t, c1.initialize("local").Error)
		require.Nil(t, c2.initialize("local").Error)

		c1.send(`{"jsonrpc":"2.0","id":2,"method":"complete","params":{"prefix":"first"}}`)
		waitCalls(t, f, 1)
		c2.send(`{"jsonrpc":"2.0","id":9,"method":"complete","params":{"prefix":"second"}}`)
		waitCalls(t, f, 2)

		r1 := c1.recv()
		require.NotNil(t, r1.Error)
		assert.Equal(t, codeCancelled, r1.Error.Code)

		close(f.release)
		r2 := c2.recv()
		assert.Equal(t, json.RawMessage("9"), r2.ID)
		assert.Contains(t, resultJSON(t, r2), "done:second:0")
	})

	t.Run("requests on different profiles do not cancel each other", func(t *testing.T) {
		f := &fakeCompleter{release: make(chan struct{})}
		s := newTestServer(t, f)
		c1, c2 := connect(t, s), connect(t, s)
		require.Nil(t, c1.initialize("local").Error)
		require.Nil(t, c2.initialize("other").Error)

		c1.send(`{"jsonrpc":"2.0","id":2,"method":"complete","params":{"prefix":"first"}}`)
		waitCalls(t, f, 1)
		c2.send(`{"jsonrpc":"2.0","id":3,"method":"complete","params":{"prefix":"second"}}`)
		waitCalls(t, f, 2)
		close(f.release)

		for _, c := range []*client{c1, c2} {
			r := c.recv()
			require.Nil(t, r.Error)
		}
	})

	t.Run("closing the connection cancels its request", func(t *testing.T) {
		f := &fakeCompleter{release: make(chan struct{})}
		s := newTestServer(t, f)
		c := connect(t, s)
		require.Nil(t, c.initialize("local").Error)
		c.send(`{"jsonrpc":"2.0","id":2,"method":"complete","params":{"prefix":"a"}}`)
		waitCalls(t, f, 1)
		c.conn.Close()

		// A following request on a fresh connection must not be blocked
		// by the orphaned one.
		c2 := connect(t, s)
		require.Nil(t, c2.initialize("local").Error)
		c2.send(`{"jsonrpc":"2.0","id":5,"method":"complete","params":{"prefix":"b"}}`)
		waitCalls(t, f, 2)
		close(f.release)
		require.Nil(t, c2.recv().Error)
	})
}

func waitCalls(t *testing.T, f *fakeCompleter, n int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		f.mu.Lock()
		got := len(f.calls)
		f.mu.Unlock()
		if got >= n {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("completer was not called %d times", n)
}

func TestInitialize_ReloadsConfig(t *testing.T) {
	conf := testConfig()
	var built []config.CompleteProfile
	s := New(Options{
		Config:       conf,
		ReloadConfig: func() (*config.Config, error) { return conf, nil },
		Version:      "test",
		NewCompleter: func(p config.CompleteProfile) complete.Completer {
			built = append(built, p)
			return &fakeCompleter{}
		},
	})
	t.Cleanup(s.Shutdown)

	require.Nil(t, connect(t, s).initialize("local").Error)
	require.Nil(t, connect(t, s).initialize("local").Error)
	assert.Len(t, built, 1, "unchanged profile reuses the backend")

	p := conf.Complete.Profiles["local"]
	p.Candidates = 3
	conf.Complete.Profiles["local"] = p
	r := connect(t, s).initialize("local")
	assert.Contains(t, resultJSON(t, r), `"candidates":3`)
	require.Len(t, built, 2, "changed profile rebuilds the backend")
	assert.Equal(t, 3, built[1].Candidates)
}
