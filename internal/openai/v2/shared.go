package openai

// TODO(micheam): Enable to support Image Content.
// TODO(micheam): Enable to support File Content.
type Message struct {
	// role string Required
	//
	// The role of the author of this message. One of system, user, or assistant.
	Role Role `json:"role"`

	// content string Required
	//
	// The contents of the message.
	Content string `json:"content"`

	// name string Optional
	//
	// The name of the author of this message. May contain a-z, A-Z, 0-9, and underscores,
	// with a maximum length of 64 characters.
	Name string `json:"name,omitempty"`
}

// SystemMessage creates a new system message
func SystemMessage(content string) *Message {
	return &Message{
		Role:    RoleSystem,
		Content: content,
	}
}

// UserMessage creates a new user message
func UserMessage(content string) *Message {
	return &Message{
		Role:    RoleUser,
		Content: content,
	}
}

// AssistantMessage creates a new assistant message
func AssistantMessage(content string) *Message {
	return &Message{
		Role:    RoleAssistant,
		Content: content,
	}
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
	RoleSystem Role = iota
	RoleUser
	RoleAssistant
)

func (r Role) String() string {
	return [...]string{"system", "user", "assistant"}[r]
}

func ParseRole(s string) Role {
	switch s {
	case "system":
		return RoleSystem
	case "user":
		return RoleUser
	case "assistant":
		return RoleAssistant
	}
	return RoleSystem
}

func (r Role) MarshalJSON() ([]byte, error) {
	return []byte(`"` + r.String() + `"`), nil
}

func (r *Role) UnmarshalJSON(b []byte) error {
	switch string(b) {
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
	Message      *Message      `json:"message,omitempty"`
	Delta        *DeltaMessage `json:"delta,omitempty"`
	FinishReason string        `json:"finish_reason"`
	Index        int           `json:"index"`
}
