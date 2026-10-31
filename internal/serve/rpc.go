package serve

import (
	"encoding/json"
	"fmt"
)

// ProtocolVersion is the version of the editor-facing protocol this server
// speaks. Clients send theirs in initialize and get 1001 on a mismatch.
const ProtocolVersion = 1

// JSON-RPC 2.0 error codes used by the server.
const (
	codeParseError     = -32700
	codeInvalidRequest = -32600
	codeMethodNotFound = -32601
	codeInvalidParams  = -32602
	codeInternalError  = -32603
	codeNotInitialized = -32002
	codeCancelled      = -32800

	codeVersionMismatch = 1001
	codeUnknownProfile  = 1002
	codeBackendError    = 1003
)

// request is an incoming JSON-RPC message: a request when ID is set, a
// notification otherwise.
type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// isNotification reports whether the message carries no id (JSON null
// counts as an id, as the spec allows, so only an absent field is a
// notification).
func (r *request) isNotification() bool { return len(r.ID) == 0 }

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

// rpcError is a JSON-RPC error object. It also implements error so handlers
// can return it directly.
type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (e *rpcError) Error() string { return fmt.Sprintf("rpc error %d: %s", e.Code, e.Message) }

func newError(code int, format string, args ...any) *rpcError {
	return &rpcError{Code: code, Message: fmt.Sprintf(format, args...)}
}

// Method names.
const (
	methodInitialize = "initialize"
	methodShutdown   = "shutdown"
	methodComplete   = "complete"
	methodCancel     = "$/cancelRequest"
)

type initializeParams struct {
	ProtocolVersion int        `json:"protocol_version"`
	Client          clientInfo `json:"client"`
	Profile         string     `json:"profile"`
}

type clientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type initializeResult struct {
	ProtocolVersion int         `json:"protocol_version"`
	Server          clientInfo  `json:"server"`
	Methods         []string    `json:"methods"`
	Profile         profileInfo `json:"profile"`
}

type profileInfo struct {
	Name        string `json:"name"`
	Model       string `json:"model"`
	Endpoint    string `json:"endpoint"`
	PrefixLines int    `json:"prefix_lines"`
	SuffixLines int    `json:"suffix_lines"`
	Candidates  int    `json:"candidates"`
}

type cancelParams struct {
	ID json.RawMessage `json:"id"`
}
