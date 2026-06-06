package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

type ChunkHandler interface {
	HandleChunk(chunk string)
	Close() error
}

// ConsoleLineStreamWriter は、改行ごとにデータを処理する io.WriteCloser です。
// 主に、コンソールへの基本的な出力を目的としています。
//
// TODO(micheam): Do we need to support io.Closer interface as well??
type ConsoleLineStreamWriter struct {
	out io.Writer
	b   bytes.Buffer
	mu  sync.Mutex
}

// HandleChunk は、受け取ったチャンクを処理します。
// ConsoleLineStreamWriter は、改行ごとにデータを処理するため、チャンクをバッファに追加し、改行があれば出力します。
func (w *ConsoleLineStreamWriter) HandleChunk(chunk string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// チャンクをバッファに追加します。
	w.b.WriteString(chunk)

	// バッファに改行があるか確認し、あれば出力します。
	for {
		line, err := w.b.ReadString('\n')
		if err != nil {
			// 改行がない場合は、バッファに残しておきます。
			w.b.WriteString(line) // 読み取れなかった部分はバッファに書き戻す
			break
		}
		fmt.Fprint(w.out, line) // 改行を含む行を出力します。
	}
}

// Close は、残っているデータを処理し、リソースを解放します。
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

// JSONLineStreamWriter は、改行ごとにデータを処理する io.WriteCloser です。
// 改行ごとのデータを JSONL 形式で出力することを目的としています。
type JSONLineStreamWriter struct {
	b   bytes.Buffer
	mu  sync.Mutex
	enc *json.Encoder

	metaData map[string]interface{} // JSONL の各行に追加するメタデータ
}

type jsonlModel struct {
	Content string `json:"content"`
	Session string `json:"session"`
	Model   string `json:"model"`
}

// HandleChunk は、受け取ったチャンクを処理します。
// JSONLineStreamWriter は、改行ごとにデータを処理するため、チャンクをバッファに追加し、改行があれば JSONL 形式で出力します。
func (w *JSONLineStreamWriter) HandleChunk(chunk string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// チャンクをバッファに追加します。
	w.b.WriteString(chunk)

	// バッファに改行があるか確認し、あれば JSONL 形式で出力します。
	for {
		line, err := w.b.ReadString('\n')
		if err != nil {
			// 改行がない場合は、バッファに残しておきます。
			w.b.WriteString(line) // 読み取れなかった部分はバッファに書き戻す
			break
		}
		// 改行を含む行を JSONL 形式で出力します。
		// jsonLine := fmt.Sprintf(`{"data": %q, "meta": %v}`, line, w.metaData)
		w.enc.Encode(jsonlModel{
			Content: line,
			Session: w.metaData["session"].(string),
			Model:   w.metaData["model"].(string),
		})
	}
}

// Close は、残っているデータを処理し、リソースを解放します。
func (w *JSONLineStreamWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	defer w.b.Reset() // バッファをクリアしてリソースを解放します。

	// バッファに残っているデータを処理します。
	if w.b.Len() > 0 {
		w.enc.Encode(jsonlModel{
			Content: w.b.String(),
			Session: w.metaData["session"].(string),
			Model:   w.metaData["model"].(string),
		})
	}
	return nil
}
