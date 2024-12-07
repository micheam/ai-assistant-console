package openai

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
)

// Message is a message in the Chat API
//
//	> A list of messages comprising the conversation so far. Depending on the model you use,
//	> different message types (modalities) are supported, like text, images, and audio.
//	>
//	> https://platform.openai.com/docs/api-reference/chat/create#chat-create-messages
//
// Message can be one of the following:
//   - System message
//   - User message
//   - Assistant message
//   - Tool message
type Message interface {
	Type() string
	json.Marshaler
	json.Unmarshaler
}

type SystemMessage struct {
	Content string
	Name    *string
}

type UserMessage struct {
	Content []Content
	Name    *string
}

type AssistantMessage struct {
	Content []Content
	Name    *string
}

var _ Message = (*SystemMessage)(nil)
var _ Message = (*UserMessage)(nil)
var _ Message = (*AssistantMessage)(nil)

func (m UserMessage) Type() string      { return "user" }
func (m SystemMessage) Type() string    { return "system" }
func (m AssistantMessage) Type() string { return "assistant" }

func (msg SystemMessage) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"role":    RoleSystem,
		"content": msg.Content,
	}
	if n := msg.Name; n != nil {
		m["name"] = *n
	}
	return json.Marshal(m)
}

func (msg UserMessage) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"role":    RoleUser,
		"content": msg.Content,
	}
	if n := msg.Name; n != nil {
		m["name"] = *n
	}
	return json.Marshal(m)
}

func (msg AssistantMessage) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"role":    RoleAssistant,
		"content": msg.Content,
	}
	if n := msg.Name; n != nil {
		m["name"] = *n
	}
	return json.Marshal(m)
}

func (msg *SystemMessage) UnmarshalJSON(b []byte) error {
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	msg.Content = m["content"].(string)
	if n, ok := m["name"].(string); ok {
		msg.Name = &n
	}
	return nil
}

func (msg *UserMessage) UnmarshalJSON(b []byte) error {
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	// NOTE: 'content' will be an array of strings or single string.
	//       We need to handle both cases.
	v := reflect.ValueOf(m["content"])
	switch v.Kind() {
	case reflect.String:
		msg.Content = []Content{&TextContent{Text: v.String()}}
	case reflect.Slice:
		msg.Content = make([]Content, v.Len())
		for i := 0; i < v.Len(); i++ {
			msg.Content = append(msg.Content, &TextContent{Text: v.Index(i).String()})
		}
	default:
		return fmt.Errorf("unexpected type for 'content': %T", m["content"])
	}
	if n, ok := m["name"].(string); ok {
		msg.Name = &n
	}
	return nil
}

func (msg *AssistantMessage) UnmarshalJSON(b []byte) error {
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	// NOTE: 'content' will be an array of strings or single string.
	//       We need to handle both cases.
	v := reflect.ValueOf(m["content"])
	switch v.Kind() {
	case reflect.String:
		msg.Content = []Content{&TextContent{Text: v.String()}}
	case reflect.Slice:
		msg.Content = make([]Content, v.Len())
		for i := 0; i < v.Len(); i++ {
			msg.Content = append(msg.Content, &TextContent{Text: v.Index(i).String()})
		}
	default:
		return fmt.Errorf("unexpected type for 'content': %T", m["content"])
	}
	if n, ok := m["name"].(string); ok {
		msg.Name = &n
	}
	return nil
}

type DeltaMessage struct {
	// content string Required
	//
	// The contents of the message.
	Content string `json:"content"`
}

// The role of the author of this message.
// One of system, user, or assistant.
type Role int

const (
	RoleUndefined Role = iota
	RoleSystem
	RoleUser
	RoleAssistant
)

var _ json.Marshaler = RoleSystem
var _ json.Unmarshaler = (*Role)(nil)

func (r Role) String() string {
	return [...]string{"undefined", "system", "user", "assistant"}[r]
}

func ParseRole(s string) Role {
	switch s {
	default:
		return RoleUndefined
	case "system":
		return RoleSystem
	case "user":
		return RoleUser
	case "assistant":
		return RoleAssistant
	}
}

func (r Role) MarshalJSON() ([]byte, error) {
	return []byte(`"` + r.String() + `"`), nil
}

func (r *Role) UnmarshalJSON(b []byte) error {
	switch string(b) {
	default:
		*r = RoleUndefined
	case `"system"`:
		*r = RoleSystem
	case `"user"`:
		*r = RoleUser
	case `"assistant"`:
		*r = RoleAssistant
	}
	return nil
}

// Usage is the usage of the Chat API
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Choice is the choice of the Chat API
type Choice struct {
	Message      AssistantMessage `json:"message,omitempty"`
	Delta        *DeltaMessage    `json:"delta,omitempty"`
	FinishReason string           `json:"finish_reason"`
	Index        int              `json:"index"`
}

// ------------------------------------------------------------------------------------------------
// Content Value of the message
// ------------------------------------------------------------------------------------------------

type Content interface {
	Type() string
	json.Marshaler
	json.Unmarshaler
}

// Variations of the content:

type TextContent struct{ Text string }
type ImageContent struct{ URL url.URL }
type AudioContent struct {
	Data   []byte // base64 encoded audio data
	Format string // wav, mp3
}
type RefusalContent struct{ Refusal string }

var _ Content = (*TextContent)(nil)
var _ Content = (*ImageContent)(nil)
var _ Content = (*AudioContent)(nil)
var _ Content = (*RefusalContent)(nil)

func (c TextContent) Type() string    { return "text" }
func (c ImageContent) Type() string   { return "image" }
func (c AudioContent) Type() string   { return "audio" }
func (c RefusalContent) Type() string { return "refusal" }

func (c TextContent) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"type": c.Type(),
		"text": c.Text,
	})
}

func (c ImageContent) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"type":    c.Type(),
		"content": c.URL.String(),
	})
}

func (c AudioContent) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type": c.Type(),
		"input_audio": map[string]string{
			"data":   string(c.Data),
			"format": c.Format,
		},
	})
}

func (c RefusalContent) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{
		"type":    c.Type(),
		"content": c.Refusal,
	})
}

func (c *TextContent) UnmarshalJSON(b []byte) error {
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	c.Text = m["text"]
	return nil
}

func (c *ImageContent) UnmarshalJSON(b []byte) error {
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	u, err := url.Parse(m["content"])
	if err != nil {
		return err
	}
	c.URL = *u
	return nil
}

func (c *AudioContent) UnmarshalJSON(b []byte) error {
	var m map[string]map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	c.Data = []byte(m["input_audio"]["data"])
	c.Format = m["input_audio"]["format"]
	return nil
}

func (c *RefusalContent) UnmarshalJSON(b []byte) error {
	var m map[string]string
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	c.Refusal = m["content"]
	return nil
}
