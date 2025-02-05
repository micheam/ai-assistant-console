package chat

import "fmt"

type Model int

// Models that are compatible with the Chat API
// https://platform.openai.com/docs/models/model-endpoint-compatibility#model-endpoint-compatibility
const (
	ModelUnspecified Model = iota
	ModelO3Mini
	ModelO1
	ModelO1Mini
	ModelGPT4o

	// GPT-4o-mini
	//
	// GPT-4o mini (“o” for “omni”) is a fast, affordable small model for focused tasks.
	// It accepts both text and image inputs, and produces text outputs (including Structured
	// Outputs). It is ideal for fine-tuning, and model outputs from a larger model like
	// GPT-4o can be distilled to GPT-4o-mini to produce similar results at lower cost and
	// latency. (see: https://platform.openai.com/docs/models/gpt-4o-mini#gpt-4o-mini)
	ModelGPT4oMini

	ModelChatGPT4oLatest
)

const DefaultModel = ModelChatGPT4oLatest

func AvailableModels() []Model {
	return []Model{
		ModelO3Mini,
		ModelO1,
		ModelO1Mini,
		ModelGPT4o,
		ModelGPT4oMini,
		ModelChatGPT4oLatest,
	}
}

func ParseModel(s string) (Model, error) {
	switch s {
	case "o3-mini":
		return ModelO3Mini, nil
	case "o1":
		return ModelO1, nil
	case "o1-mini":
		return ModelO1Mini, nil
	case "gpt-4o":
		return ModelGPT4o, nil
	case "gpt-4o-mini":
		return ModelGPT4oMini, nil
	case "chatgpt-4o-latest":
		return ModelChatGPT4oLatest, nil
	default:
		return ModelUnspecified, fmt.Errorf("unsupported model: %s", s)
	}
}

func (m Model) String() string {
	switch m {
	case ModelUnspecified:
		return "<unspecified>"
	case ModelO3Mini:
		return "o3-mini"
	case ModelO1:
		return "o1"
	case ModelO1Mini:
		return "o1-mini"
	case ModelGPT4o:
		return "gpt-4o"
	case ModelGPT4oMini:
		return "gpt-4o-mini"
	case ModelChatGPT4oLatest:
		return "chatgpt-4o-latest"
	default:
		panic(fmt.Sprintf("unsupported model: %d", m))
	}
}

func (m Model) MarshalJSON() ([]byte, error) {
	return []byte(`"` + m.String() + `"`), nil
}
