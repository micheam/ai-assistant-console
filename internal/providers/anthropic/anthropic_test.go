package anthropic

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	anthropicsdk "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"
	"github.com/stretchr/testify/require"

	"micheam.com/aico/internal/assistant"
)

func TestBuildRequestBody_ToolsAndToolChoice(t *testing.T) {
	ctx := context.Background()

	t.Run("no tools: omits tools and tool_choice", func(t *testing.T) {
		body, err := buildRequestBody(ctx, anthropicsdk.Model("claude-fable-5-1"), nil, nil, assistant.GenerationOptions{}, nil)
		require.NoError(t, err)
		data, err := body.MarshalJSON()
		require.NoError(t, err)
		require.NotContains(t, string(data), `"tools"`)
		require.NotContains(t, string(data), `"tool_choice"`)
	})

	t.Run("with tools: sets tools and tool_choice auto", func(t *testing.T) {
		tools := []assistant.ToolDefinition{
			{Name: "propose_edit", Description: "propose an edit", InputSchema: map[string]any{"type": "object"}},
		}
		body, err := buildRequestBody(ctx, anthropicsdk.Model("claude-fable-5-1"), nil, tools, assistant.GenerationOptions{}, nil)
		require.NoError(t, err)
		data, err := body.MarshalJSON()
		require.NoError(t, err)
		require.Contains(t, string(data), `"tools"`)
		require.Contains(t, string(data), `"propose_edit"`)
		require.Contains(t, string(data), `"tool_choice":{"type":"auto"}`)
	})
}

func TestBuildRequestBody_MaxTokens(t *testing.T) {
	ctx := context.Background()

	t.Run("zero: keeps the default", func(t *testing.T) {
		body, err := buildRequestBody(ctx, anthropicsdk.Model("claude-opus-5-5"), nil, nil, assistant.GenerationOptions{}, nil)
		require.NoError(t, err)
		data, err := body.MarshalJSON()
		require.NoError(t, err)
		require.Contains(t, string(data), `"max_tokens":32768`)
	})

	t.Run("set: overrides the default", func(t *testing.T) {
		body, err := buildRequestBody(ctx, anthropicsdk.Model("claude-opus-5-5"), nil, nil, assistant.GenerationOptions{MaxTokens: 65536}, nil)
		require.NoError(t, err)
		data, err := body.MarshalJSON()
		require.NoError(t, err)
		require.Contains(t, string(data), `"max_tokens":65536`)
	})
}

// TestGenerateContentStream_Effort goes through a real HTTP round trip
// because output_config.effort is injected into the serialized body at send
// time (option.WithJSONSet), so it is invisible to
// MessageNewParams.MarshalJSON. It uses the streaming path, which is the
// one the CLI exercises.
func TestGenerateContentStream_Effort(t *testing.T) {
	const reply = "event: message_start\n" +
		`data: {"type":"message_start","message":{"id":"msg_01","type":"message","role":"assistant","model":"claude-opus-5-5","content":[],"stop_reason":null,"usage":{"input_tokens":1,"output_tokens":0}}}` + "\n\n" +
		"event: content_block_start\n" +
		`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}` + "\n\n" +
		"event: content_block_delta\n" +
		`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"ok"}}` + "\n\n" +
		"event: content_block_stop\n" +
		`data: {"type":"content_block_stop","index":0}` + "\n\n" +
		"event: message_delta\n" +
		`data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":1}}` + "\n\n" +
		"event: message_stop\n" +
		`data: {"type":"message_stop"}` + "\n\n"

	run := func(t *testing.T, genOpts assistant.GenerationOptions) string {
		t.Helper()
		var got string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			got = string(b)
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte(reply))
		}))
		t.Cleanup(srv.Close)

		client := anthropicsdk.NewClient(option.WithAPIKey("test"), option.WithBaseURL(srv.URL))
		msgs := []assistant.Message{assistant.NewUserMessage(assistant.NewTextContent("hi"))}
		it, err := generateContentStream(context.Background(), client, "claude-opus-5-5", nil, nil, genOpts, msgs)
		require.NoError(t, err)
		for _, err := range it {
			require.NoError(t, err)
		}
		return got
	}

	t.Run("empty: omits output_config", func(t *testing.T) {
		body := run(t, assistant.GenerationOptions{})
		require.NotContains(t, body, `"output_config"`)
	})

	t.Run("set: sends output_config.effort and max_tokens", func(t *testing.T) {
		body := run(t, assistant.GenerationOptions{Effort: "xhigh", MaxTokens: 65536})
		require.Contains(t, body, `"output_config":{"effort":"xhigh"}`)
		require.Contains(t, body, `"max_tokens":65536`)
	})
}

func TestConvertContentBlockParamUnion(t *testing.T) {
	t.Run("ToolUseContent", func(t *testing.T) {
		c := &assistant.ToolUseContent{
			ID:    "toolu_01",
			Name:  "propose_edit",
			Input: json.RawMessage(`{"description":"x"}`),
		}
		block, err := convertContentBlockParamUnion(c)
		require.NoError(t, err)
		data, err := json.Marshal(block)
		require.NoError(t, err)
		require.Contains(t, string(data), `"type":"tool_use"`)
		require.Contains(t, string(data), `"toolu_01"`)
	})

	t.Run("ToolResultContent", func(t *testing.T) {
		c := &assistant.ToolResultContent{ToolUseID: "toolu_01", Content: "applied", IsError: false}
		block, err := convertContentBlockParamUnion(c)
		require.NoError(t, err)
		data, err := json.Marshal(block)
		require.NoError(t, err)
		require.Contains(t, string(data), `"type":"tool_result"`)
		require.Contains(t, string(data), `"toolu_01"`)
	})

	t.Run("ThinkingContent", func(t *testing.T) {
		c := &assistant.ThinkingContent{Thinking: "", Signature: "sig"}
		block, err := convertContentBlockParamUnion(c)
		require.NoError(t, err)
		data, err := json.Marshal(block)
		require.NoError(t, err)
		require.Contains(t, string(data), `"type":"thinking"`)
		require.Contains(t, string(data), `"sig"`)
	})

	t.Run("RedactedThinkingContent", func(t *testing.T) {
		c := &assistant.RedactedThinkingContent{Data: "opaque"}
		block, err := convertContentBlockParamUnion(c)
		require.NoError(t, err)
		data, err := json.Marshal(block)
		require.NoError(t, err)
		require.Contains(t, string(data), `"type":"redacted_thinking"`)
		require.Contains(t, string(data), `"opaque"`)
	})
}

// fakeDecoder feeds a fixed sequence of ssestream.Event values, matching
// the shape ssestream.Stream expects: Event.Type drives which events are
// unmarshaled ("message_start", "content_block_start", ... are kept;
// "ping" is skipped; "error" fails the stream).
type fakeDecoder struct {
	events []ssestream.Event
	i      int
}

func (f *fakeDecoder) Event() ssestream.Event { return f.events[f.i-1] }
func (f *fakeDecoder) Next() bool {
	if f.i >= len(f.events) {
		return false
	}
	f.i++
	return true
}
func (f *fakeDecoder) Close() error { return nil }
func (f *fakeDecoder) Err() error   { return nil }

func sseEvent(typ, data string) ssestream.Event {
	return ssestream.Event{Type: typ, Data: []byte(data)}
}

func newTestStream(events []ssestream.Event) *ssestream.Stream[anthropicsdk.MessageStreamEvent] {
	return ssestream.NewStream[anthropicsdk.MessageStreamEvent](&fakeDecoder{events: events}, nil)
}

func TestStreamContent_TextThenToolUseThenUsage(t *testing.T) {
	events := []ssestream.Event{
		sseEvent("message_start", `{"type":"message_start","message":{"id":"msg_1","type":"message","role":"assistant","content":[],"model":"claude-fable-5-1","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":10,"output_tokens":0}}}`),
		sseEvent("content_block_start", `{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`),
		sseEvent("content_block_delta", `{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Here is a fix:"}}`),
		sseEvent("content_block_stop", `{"type":"content_block_stop","index":0}`),
		sseEvent("content_block_start", `{"type":"content_block_start","index":1,"content_block":{"type":"tool_use","id":"toolu_01","name":"propose_edit","input":{}}}`),
		sseEvent("content_block_delta", `{"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":"{\"description\":\"x\",\"old_string\":\"a\",\"new_string\":\"b\"}"}}`),
		sseEvent("content_block_stop", `{"type":"content_block_stop","index":1}`),
		sseEvent("message_delta", `{"type":"message_delta","delta":{"stop_reason":"tool_use","stop_sequence":null},"usage":{"output_tokens":20}}`),
		sseEvent("message_stop", `{"type":"message_stop"}`),
	}

	var got []*assistant.GenerateContentResponse
	var streamErr error
	for resp, err := range streamContent(context.Background(), newTestStream(events)) {
		if err != nil {
			streamErr = err
			break
		}
		got = append(got, resp)
	}
	require.NoError(t, streamErr)
	require.Len(t, got, 3)

	text, ok := got[0].Content.(*assistant.TextContent)
	require.True(t, ok)
	require.Equal(t, "Here is a fix:", text.Text)

	tu, ok := got[1].Content.(*assistant.ToolUseContent)
	require.True(t, ok)
	require.Equal(t, "toolu_01", tu.ID)
	require.Equal(t, "propose_edit", tu.Name)
	require.JSONEq(t, `{"description":"x","old_string":"a","new_string":"b"}`, string(tu.Input))

	require.Nil(t, got[2].Content)
	require.NotNil(t, got[2].Usage)
	require.Equal(t, 20, got[2].Usage.OutputTokens)
}

func TestStreamContent_TruncatedToolUse_ReturnsErrTruncatedToolUse(t *testing.T) {
	events := []ssestream.Event{
		sseEvent("message_start", `{"type":"message_start","message":{"id":"msg_1","type":"message","role":"assistant","content":[],"model":"claude-fable-5-1","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":10,"output_tokens":0}}}`),
		sseEvent("content_block_start", `{"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_01","name":"propose_edit","input":{}}}`),
		sseEvent("content_block_delta", `{"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"description\":\"x\", \"old_str"}}`),
		sseEvent("content_block_stop", `{"type":"content_block_stop","index":0}`),
	}

	var streamErr error
	for _, err := range streamContent(context.Background(), newTestStream(events)) {
		if err != nil {
			streamErr = err
			break
		}
	}
	require.Error(t, streamErr)
	require.True(t, errors.Is(streamErr, assistant.ErrTruncatedToolUse))
}

func TestStreamContent_MaxTokens_ReturnsErrMaxTokens(t *testing.T) {
	events := []ssestream.Event{
		sseEvent("message_start", `{"type":"message_start","message":{"id":"msg_1","type":"message","role":"assistant","content":[],"model":"claude-fable-5-1","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":10,"output_tokens":0}}}`),
		sseEvent("content_block_start", `{"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}`),
		sseEvent("content_block_delta", `{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"partial"}}`),
		sseEvent("content_block_stop", `{"type":"content_block_stop","index":0}`),
		sseEvent("message_delta", `{"type":"message_delta","delta":{"stop_reason":"max_tokens","stop_sequence":null},"usage":{"output_tokens":8192}}`),
		sseEvent("message_stop", `{"type":"message_stop"}`),
	}

	var streamErr error
	for _, err := range streamContent(context.Background(), newTestStream(events)) {
		if err != nil {
			streamErr = err
			break
		}
	}
	require.Error(t, streamErr)
	require.True(t, errors.Is(streamErr, assistant.ErrMaxTokens))
}

func TestStreamContent_Refusal_ReturnsErrRefusal(t *testing.T) {
	events := []ssestream.Event{
		sseEvent("message_start", `{"type":"message_start","message":{"id":"msg_1","type":"message","role":"assistant","content":[],"model":"claude-fable-5-1","stop_reason":null,"stop_sequence":null,"usage":{"input_tokens":10,"output_tokens":0}}}`),
		sseEvent("message_delta", `{"type":"message_delta","delta":{"stop_reason":"refusal","stop_sequence":null},"usage":{"output_tokens":0}}`),
		sseEvent("message_stop", `{"type":"message_stop"}`),
	}

	var streamErr error
	for _, err := range streamContent(context.Background(), newTestStream(events)) {
		if err != nil {
			streamErr = err
			break
		}
	}
	require.Error(t, streamErr)
	require.True(t, errors.Is(streamErr, assistant.ErrRefusal))
}

func TestAliases_PointToSupportedModels(t *testing.T) {
	for alias, target := range Aliases() {
		t.Run(alias, func(t *testing.T) {
			m, ok := selectModel(target)
			require.True(t, ok, "alias %q points to unknown model %q", alias, target)
			require.False(t, strings.HasPrefix(m.Description(), "[Deprecated]"),
				"alias %q points to deprecated model %q", alias, target)
		})
	}
}
