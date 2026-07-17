package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"micheam.com/aico/internal/assistant"
)

type ConsoleLineStreamWriter struct {
	out io.Writer
	b   bytes.Buffer
	mu  sync.Mutex
}

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

// SetUsage attaches usage info to be emitted with the next flushed line.
// Intended to be called once, after the generation stream has completed.
func (w *JSONLineStreamWriter) SetUsage(usage *assistant.Usage) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.usage = usage
}

type jsonlModel struct {
	Content string     `json:"content"`
	Session string     `json:"session"`
	Model   string     `json:"model"`
	Usage   *usageInfo `json:"usage,omitempty"`
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
