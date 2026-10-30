package serve

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadMessage(t *testing.T) {
	tests := map[string]struct {
		input   string
		want    []string
		wantErr string
	}{
		"frame as Vim's lsp mode sends it": {
			input: "Content-Length: 65\r\n\r\n" + `{"method":"fim.complete","jsonrpc":"2.0","id":1,"params":{"a":1}}`,
			want:  []string{`{"method":"fim.complete","jsonrpc":"2.0","id":1,"params":{"a":1}}`},
		},
		"two consecutive frames": {
			input: "Content-Length: 2\r\n\r\n{}Content-Length: 3\r\n\r\n[1]",
			want:  []string{"{}", "[1]"},
		},
		"extra header and lowercase name": {
			input: "content-type: application/vscode-jsonrpc; charset=utf-8\r\ncontent-length: 2\r\n\r\n{}",
			want:  []string{"{}"},
		},
		"missing Content-Length": {
			input:   "Content-Type: x\r\n\r\n{}",
			wantErr: "missing Content-Length",
		},
		"invalid Content-Length": {
			input:   "Content-Length: abc\r\n\r\n{}",
			wantErr: "invalid Content-Length",
		},
		"body shorter than declared": {
			input:   "Content-Length: 10\r\n\r\n{}",
			wantErr: "read body",
		},
		"over size limit": {
			input:   "Content-Length: 99999999999\r\n\r\n",
			wantErr: errMessageTooLarge.Error(),
		},
		"malformed header line": {
			input:   "garbage\r\n\r\n{}",
			wantErr: "malformed header",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(tt.input))
			var got []string
			for {
				body, err := readMessage(r)
				if err == io.EOF {
					break
				}
				if tt.wantErr != "" {
					require.ErrorContains(t, err, tt.wantErr)
					return
				}
				require.NoError(t, err)
				got = append(got, string(body))
			}
			require.Empty(t, tt.wantErr, "expected an error")
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReadMessage_EOF(t *testing.T) {
	_, err := readMessage(bufio.NewReader(strings.NewReader("")))
	assert.Equal(t, io.EOF, err)
}

func TestWriteMessage_RoundTrip(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, writeMessage(&buf, []byte(`{"id":1}`)))
	assert.Equal(t, "Content-Length: 8\r\n\r\n"+`{"id":1}`, buf.String())

	body, err := readMessage(bufio.NewReader(&buf))
	require.NoError(t, err)
	assert.Equal(t, `{"id":1}`, string(body))
}
