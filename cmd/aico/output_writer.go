package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sync"
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
}
type JSONOutputMetaData struct {
	Session string
	Model   string
}

type jsonlModel struct {
	Content string `json:"content"`
	Session string `json:"session"`
	Model   string `json:"model"`
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
	if w.b.Len() == 0 {
		return nil
	}
	return w.enc.Encode(jsonlModel{
		Content: w.b.String(),
		Session: w.metaData.Session,
		Model:   w.metaData.Model,
	})
}
