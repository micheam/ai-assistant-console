package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"micheam.com/aico/internal/assistant"
)

// streamWriter is the interface doGenerate consumes to render a
// generation's output. In addition to the plain io.WriteCloser text
// stream, it can emit a tool_use content block as a distinct record so it
// can be told apart from ordinary text on the wire (JSONL) or rendered
// distinctly for a human reader (console).
type streamWriter interface {
	io.WriteCloser
	EmitToolUse(*assistant.ToolUseContent) error
}

type ConsoleLineStreamWriter struct {
	out io.Writer
	b   bytes.Buffer
	mu  sync.Mutex
}

var _ streamWriter = (*ConsoleLineStreamWriter)(nil)

func (w *ConsoleLineStreamWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	n, err = w.b.Write(p)
	if err != nil {
		return n, err
	}

	for {
		line, err := w.b.ReadBytes('\n')
		if err != nil {
			// 改行がない場合は、バッファに残しておきます。
			w.b.Write(line) // 読み取れなかった部分はバッファに書き戻す
			break
		}
		w.out.Write(line) // 改行ごとに出力します。
	}
	return n, nil
}

func (w *ConsoleLineStreamWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	defer w.b.Reset() // バッファをクリアしてリソースを解放します。

	// バッファに残っているデータを処理します。
	if w.b.Len() > 0 {
		fmt.Fprint(w.out, w.b.String())
	}
	return nil
}

// EmitToolUse flushes any buffered partial line (so it is not interleaved
// with the tool_use block out of order) and prints a human-readable
// rendering of the proposal.
func (w *ConsoleLineStreamWriter) EmitToolUse(tu *assistant.ToolUseContent) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.b.Len() > 0 {
		fmt.Fprint(w.out, w.b.String())
		if !strings.HasSuffix(w.b.String(), "\n") {
			fmt.Fprintln(w.out)
		}
		w.b.Reset()
	}

	if tu.Name != toolNameProposeEdit {
		fmt.Fprintf(w.out, "[tool_use %s] %s\n", tu.Name, string(tu.Input))
		return nil
	}
	description, oldString, newString, ok := proposeEditFields(tu.Input)
	if !ok {
		fmt.Fprintf(w.out, "[propose_edit] %s\n", string(tu.Input))
		return nil
	}
	fmt.Fprintf(w.out, "[propose_edit] %s\n", description)
	for _, l := range strings.Split(oldString, "\n") {
		fmt.Fprintf(w.out, "-%s\n", l)
	}
	for _, l := range strings.Split(newString, "\n") {
		fmt.Fprintf(w.out, "+%s\n", l)
	}
	return nil
}

type JSONLineStreamWriter struct {
	b        bytes.Buffer
	mu       sync.Mutex
	enc      *json.Encoder
	metaData JSONOutputMetaData
	usage    *assistant.Usage
}
type JSONOutputMetaData struct {
	Session string
	Model   string
}

var _ streamWriter = (*JSONLineStreamWriter)(nil)

// SetUsage attaches usage info to be emitted with the next flushed line.
// Intended to be called once, after the generation stream has completed.
func (w *JSONLineStreamWriter) SetUsage(usage *assistant.Usage) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.usage = usage
}

type jsonlModel struct {
	Content string       `json:"content"`
	ToolUse *toolUseView `json:"tool_use,omitempty"`
	Session string       `json:"session"`
	Model   string       `json:"model"`
	Usage   *usageInfo   `json:"usage,omitempty"`
}

// toolUseView is the JSONL-facing shape of a tool_use content block.
type toolUseView struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

// usageInfo is the JSON-facing shape of assistant.Usage, adding a
// human-readable cache hit rate alongside the raw token counts.
type usageInfo struct {
	InputTokens       int    `json:"input_tokens"`
	OutputTokens      int    `json:"output_tokens"`
	CachedInputTokens int    `json:"cached_input_tokens"`
	CacheWriteTokens  int    `json:"cache_write_tokens"`
	CacheHitRate      string `json:"cache_hit_rate"`
}

func toUsageInfo(u *assistant.Usage) *usageInfo {
	if u == nil {
		return nil
	}
	return &usageInfo{
		InputTokens:       u.InputTokens,
		OutputTokens:      u.OutputTokens,
		CachedInputTokens: u.CachedInputTokens,
		CacheWriteTokens:  u.CacheWriteTokens,
		CacheHitRate:      fmt.Sprintf("%.1f%%", u.CacheHitRate()),
	}
}

func (w *JSONLineStreamWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	n, err = w.b.Write(p)
	if err != nil {
		return n, err
	}
	for {
		line, err := w.b.ReadBytes('\n')
		if err != nil {
			// 改行がない場合は、バッファに残しておきます。
			w.b.Write(line)
			break
		}
		if err := w.enc.Encode(jsonlModel{
			Content: string(line),
			Session: w.metaData.Session,
			Model:   w.metaData.Model,
		}); err != nil {
			return n, err
		}
	}
	return n, nil
}

// EmitToolUse flushes any buffered partial line as its own content record
// (preserving output order), then emits a tool_use record. Content is
// always present (empty on the tool_use record itself) so a client that
// only reads jsonlModel.Content — such as older vim-aico builds — ignores
// tool_use records harmlessly instead of breaking on them.
func (w *JSONLineStreamWriter) EmitToolUse(tu *assistant.ToolUseContent) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.b.Len() > 0 {
		if err := w.enc.Encode(jsonlModel{
			Content: w.b.String(),
			Session: w.metaData.Session,
			Model:   w.metaData.Model,
		}); err != nil {
			return err
		}
		w.b.Reset()
	}

	input := tu.Input
	if len(input) == 0 {
		input = json.RawMessage("{}")
	}
	return w.enc.Encode(jsonlModel{
		Content: "",
		ToolUse: &toolUseView{ID: tu.ID, Name: tu.Name, Input: input},
		Session: w.metaData.Session,
		Model:   w.metaData.Model,
	})
}

func (w *JSONLineStreamWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	defer w.b.Reset()
	if w.b.Len() == 0 && w.usage == nil {
		return nil
	}
	return w.enc.Encode(jsonlModel{
		Content: w.b.String(),
		Session: w.metaData.Session,
		Model:   w.metaData.Model,
		Usage:   toUsageInfo(w.usage),
	})
}
