package openai

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
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
	Index        int     `json:"index"`
}
