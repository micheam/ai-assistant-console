package serve

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"sync"

	"micheam.com/aico/internal/complete"
	"micheam.com/aico/internal/config"
)

// Options configures a Server.
type Options struct {
	// Config supplies the completion profiles.
	Config *config.Config

	// ReloadConfig, when set, is called on every initialize so profile
	// changes in the config file take effect without restarting the
	// daemon. On error the last good config is kept.
	ReloadConfig func() (*config.Config, error)

	// Version is reported to clients in initialize.
	Version string

	// NewCompleter builds the backend client for a profile. It is called
	// once per profile, on the first initialize that selects it.
	NewCompleter func(config.CompleteProfile) complete.Completer

	// Logger receives operational logs. Nil discards them.
	Logger *slog.Logger
}

// Server accepts editor connections and serves completion requests.
//
// At most one completion request runs per profile: a new request cancels
// the one in flight, whichever connection it came from.
type Server struct {
	opts Options
	log  *slog.Logger

	mu       sync.Mutex
	profiles map[string]*profileState
	conns    map[*conn]struct{}
	listener net.Listener
	closed   bool
	done     chan struct{}
}

// profileState is what the server keeps per profile across connections.
type profileState struct {
	profile   config.CompleteProfile // settings the completer was built from
	completer complete.Completer

	mu       sync.Mutex
	inflight *inflight
}

// inflight identifies one running completion request.
type inflight struct {
	cancel context.CancelFunc
}

// New returns a Server for the given options.
func New(opts Options) *Server {
	log := opts.Logger
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	return &Server{
		opts:     opts,
		log:      log,
		profiles: map[string]*profileState{},
		conns:    map[*conn]struct{}{},
		done:     make(chan struct{}),
	}
}

// Serve accepts connections on ln until Shutdown is called or ln fails.
// It returns nil after Shutdown.
func (s *Server) Serve(ln net.Listener) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		ln.Close()
		return errors.New("server is shut down")
	}
	s.listener = ln
	s.mu.Unlock()

	for {
		c, err := ln.Accept()
		if err != nil {
			select {
			case <-s.done:
				return nil
			default:
				return err
			}
		}
		go s.ServeConn(c)
	}
}

// Shutdown stops accepting, closes every connection and makes Serve
// return. It is safe to call more than once.
func (s *Server) Shutdown() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	close(s.done)
	if s.listener != nil {
		s.listener.Close()
	}
	for c := range s.conns {
		c.rw.Close()
	}
}

// Done is closed once Shutdown has been called.
func (s *Server) Done() <-chan struct{} { return s.done }

// ServeConn serves one client on rw until it closes or the server shuts
// down. It is exported so tests can drive the server over an in-memory
// pipe.
func (s *Server) ServeConn(rw io.ReadWriteCloser) {
	c := &conn{server: s, rw: rw, pending: map[string]context.CancelFunc{}}

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		rw.Close()
		return
	}
	s.conns[c] = struct{}{}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.conns, c)
		s.mu.Unlock()
		rw.Close()
		c.cancelAll()
	}()
	c.readLoop()
}

// profile returns the shared state for the named profile, creating the
// backend client on first use and rebuilding it when the profile's
// settings changed since.
func (s *Server) profile(name string) (string, config.CompleteProfile, *profileState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.opts.ReloadConfig != nil {
		if conf, err := s.opts.ReloadConfig(); err != nil {
			s.log.Warn("reload config failed; keeping the previous one", "error", err)
		} else {
			s.opts.Config = conf
		}
	}
	resolved, p, err := s.opts.Config.CompleteProfile(name)
	if err != nil {
		return "", config.CompleteProfile{}, nil, err
	}
	st, ok := s.profiles[resolved]
	if !ok {
		st = &profileState{profile: p, completer: s.opts.NewCompleter(p)}
		s.profiles[resolved] = st
	} else if st.profile != p {
		s.log.Info("profile settings changed; rebuilding backend", "profile", resolved)
		st.profile, st.completer = p, s.opts.NewCompleter(p)
	}
	return resolved, p, st, nil
}

// begin registers cancel as the profile's in-flight request, cancelling
// the previous one, and returns a token for end.
func (st *profileState) begin(cancel context.CancelFunc) *inflight {
	st.mu.Lock()
	defer st.mu.Unlock()
	if st.inflight != nil {
		st.inflight.cancel()
	}
	st.inflight = &inflight{cancel: cancel}
	return st.inflight
}

// end clears the in-flight slot if tok still owns it.
func (st *profileState) end(tok *inflight) {
	st.mu.Lock()
	defer st.mu.Unlock()
	if st.inflight == tok {
		st.inflight = nil
	}
}

// conn is one client connection.
type conn struct {
	server *Server
	rw     io.ReadWriteCloser

	writeMu sync.Mutex

	mu          sync.Mutex
	initialized bool
	profileName string
	profile     config.CompleteProfile
	state       *profileState
	pending     map[string]context.CancelFunc // in-flight requests by id
}

func (c *conn) readLoop() {
	r := bufio.NewReader(c.rw)
	for {
		body, err := readMessage(r)
		if err != nil {
			if err != io.EOF {
				c.server.log.Debug("connection read error", "error", err)
				if errors.Is(err, errMessageTooLarge) {
					c.reply(nil, nil, newError(codeInvalidRequest, "%v", err))
				}
			}
			return
		}
		var req request
		if err := json.Unmarshal(body, &req); err != nil {
			c.reply(nil, nil, newError(codeParseError, "parse error: %v", err))
			continue
		}
		c.server.log.Debug("message received", "method", req.Method, "id", string(req.ID))
		if req.isNotification() {
			c.handleNotification(&req)
			continue
		}
		go c.handleRequest(&req)
	}
}

func (c *conn) handleNotification(req *request) {
	switch req.Method {
	case methodCancel:
		var p cancelParams
		if err := json.Unmarshal(req.Params, &p); err != nil || len(p.ID) == 0 {
			return
		}
		c.cancelPending(string(p.ID))
	default:
		c.server.log.Debug("ignoring unknown notification", "method", req.Method)
	}
}

func (c *conn) handleRequest(req *request) {
	var (
		result any
		err    error
	)
	switch req.Method {
	case methodInitialize:
		result, err = c.initialize(req.Params)
	case methodShutdown:
		result, err = c.shutdown()
	case methodComplete:
		result, err = c.complete(req.ID, req.Params)
	default:
		err = newError(codeMethodNotFound, "unknown method %q", req.Method)
	}
	c.reply(req.ID, result, err)
}

func (c *conn) initialize(params json.RawMessage) (any, error) {
	var p initializeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, newError(codeInvalidParams, "invalid initialize params: %v", err)
	}
	if p.ProtocolVersion != ProtocolVersion {
		return nil, &rpcError{
			Code:    codeVersionMismatch,
			Message: "unsupported protocol version",
			Data:    map[string]int{"protocol_version": ProtocolVersion},
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.initialized {
		return nil, newError(codeInvalidRequest, "already initialized")
	}
	name, profile, state, err := c.server.profile(p.Profile)
	if err != nil {
		return nil, newError(codeUnknownProfile, "%v", err)
	}
	c.initialized = true
	c.profileName, c.profile, c.state = name, profile, state
	c.server.log.Info("client initialized",
		"client", p.Client.Name, "client_version", p.Client.Version, "profile", name)

	return initializeResult{
		ProtocolVersion: ProtocolVersion,
		Server:          clientInfo{Name: "aico", Version: c.server.opts.Version},
		Methods:         []string{methodComplete},
		Profile: profileInfo{
			Name:        name,
			Model:       profile.Model,
			Endpoint:    profile.Endpoint,
			PrefixLines: profile.PrefixLines,
			SuffixLines: profile.SuffixLines,
			Candidates:  profile.Candidates,
		},
	}, nil
}

func (c *conn) shutdown() (any, error) {
	c.server.log.Info("shutdown requested")
	// Shut down after the response has been written; see reply.
	return shutdownResult{}, nil
}

// shutdownResult marks a shutdown response so reply can trigger the
// server shutdown once the response is on the wire.
type shutdownResult struct{}

func (c *conn) complete(id json.RawMessage, params json.RawMessage) (any, error) {
	c.mu.Lock()
	initialized, state, profile := c.initialized, c.state, c.profile
	c.mu.Unlock()
	if !initialized {
		return nil, newError(codeNotInitialized, "initialize first")
	}

	var req complete.Request
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, newError(codeInvalidParams, "invalid complete params: %v", err)
	}
	if req.Count < 0 {
		return nil, newError(codeInvalidParams, "count must not be negative")
	}
	if req.Count == 0 {
		req.Count = profile.Candidates
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c.addPending(string(id), cancel)
	defer c.removePending(string(id))
	tok := state.begin(cancel)
	defer state.end(tok)

	res, err := state.completer.Complete(ctx, req)
	switch {
	case ctx.Err() != nil:
		return nil, newError(codeCancelled, "request cancelled")
	case err != nil:
		c.server.log.Warn("completion failed", "profile", c.profileName, "error", err)
		return nil, newError(codeBackendError, "%v", err)
	}
	return res, nil
}

func (c *conn) addPending(id string, cancel context.CancelFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pending[id] = cancel
}

func (c *conn) removePending(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.pending, id)
}

func (c *conn) cancelPending(id string) {
	c.mu.Lock()
	cancel := c.pending[id]
	c.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (c *conn) cancelAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, cancel := range c.pending {
		cancel()
	}
}

// reply writes a response. A nil id (used for parse errors) is sent as
// JSON null, as the spec requires.
func (c *conn) reply(id json.RawMessage, result any, err error) {
	resp := response{JSONRPC: "2.0", ID: id}
	if len(id) == 0 {
		resp.ID = json.RawMessage("null")
	}
	if err != nil {
		var rerr *rpcError
		if !errors.As(err, &rerr) {
			rerr = newError(codeInternalError, "%v", err)
		}
		resp.Error = rerr
	} else {
		resp.Result = result
	}
	body, merr := json.Marshal(resp)
	if merr != nil {
		c.server.log.Error("marshal response", "error", merr)
		return
	}

	c.writeMu.Lock()
	werr := writeMessage(c.rw, body)
	c.writeMu.Unlock()
	if werr != nil {
		c.server.log.Debug("write response", "error", werr)
	}
	if _, ok := result.(shutdownResult); ok {
		c.server.Shutdown()
	}
}
