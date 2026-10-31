// Package serve implements the `aico serve` daemon: a JSON-RPC 2.0 server
// over a Unix domain socket that editor clients use for code completion.
//
// Messages are framed like Language Server Protocol messages, so Vim's
// channel "lsp" mode can talk to it directly.
package serve

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// maxMessageSize caps a single message body. Editor clients send whole
// buffers as context, so this is generous, but it still bounds memory
// against a misbehaving client.
const maxMessageSize = 64 << 20

var errMessageTooLarge = errors.New("message body exceeds size limit")

// readMessage reads one framed message body from r.
//
// The frame is one or more "Name: value" header lines, each ending in
// "\r\n", followed by an empty line and Content-Length bytes of body.
// Header names are matched case-insensitively; headers other than
// Content-Length are ignored. It returns io.EOF when the stream ends
// cleanly before a frame starts.
func readMessage(r *bufio.Reader) ([]byte, error) {
	length := -1
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF && line == "" && length < 0 {
				return nil, io.EOF
			}
			return nil, fmt.Errorf("read header: %w", err)
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		name, value, ok := strings.Cut(line, ":")
		if !ok {
			return nil, fmt.Errorf("malformed header line %q", line)
		}
		if !strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
			continue
		}
		n, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil || n < 0 {
			return nil, fmt.Errorf("invalid Content-Length %q", strings.TrimSpace(value))
		}
		length = n
	}
	if length < 0 {
		return nil, errors.New("missing Content-Length header")
	}
	if length > maxMessageSize {
		return nil, errMessageTooLarge
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	return body, nil
}

// writeMessage writes body to w as one framed message.
func writeMessage(w io.Writer, body []byte) error {
	if _, err := fmt.Fprintf(w, "Content-Length: %d\r\n\r\n", len(body)); err != nil {
		return err
	}
	_, err := w.Write(body)
	return err
}
