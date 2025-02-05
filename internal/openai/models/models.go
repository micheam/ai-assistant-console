package models

type Model string

const (
	// GPT-4o (“o” for “omni”) is a versatile, high-intelligence flagship model of OpenAI.
	// It supports both text and image inputs and generates text outputs, including structured responses.
	// The model ID "chatgpt-4o-latest" dynamically points to the latest version of GPT-4o used in ChatGPT
	// and is updated when significant changes occur.
	// The knowledge cutoff date for GPT-4o models is October 2023.
	// Reference: https://platform.openai.com/docs/models#gpt-4o
	GPT4O Model = "gpt-4o"

	// GPT-4o-mini (“o” for “omni”) is a fast, affordable small model for focused tasks.
	// It accepts both text and image inputs, and produces text outputs, including structured outputs.
	// Outputs from a larger model like GPT-4o can be distilled into GPT-4o-mini to achieve similar
	// results with reduced cost and latency.
	// The knowledge cutoff date for GPT-4o-mini models is October 2023.
	// Reference: https://platform.openai.com/docs/models#gpt-4o-mini
	GPT4OMini Model = "gpt-4o-mini"
)

// The o1 series of models are trained with reinforcement learning for complex reasoning tasks.
// These models generate a long internal chain of thought before responding.
// o1-mini is a faster and more affordable reasoning model, but o3-mini is recommended for higher intelligence
// at the same cost and latency.
// The latest o1 model supports both text and image inputs, while o1-mini supports text inputs only.
// The knowledge cutoff date for o1 and o1-mini models is October 2023.
// Reference: https://platform.openai.com/docs/models#o1
const (
	O1     Model = "o1"
	O1Mini Model = "o1-mini"
)

const (
	// o3-mini is the most recent small reasoning model, providing high intelligence at the same cost and latency as o1-mini.
	// It supports key developer features, including structured outputs, function calling, and the Batch API.
	// Like other models in the o-series, it is optimised for science, math, and coding tasks.
	// The knowledge cutoff date for o3-mini models is October 2023.
	// Reference: https://platform.openai.com/docs/models#o3-mini
	O3Mini Model = "o3-mini"
)

// DALL·E is an AI system that generates images and art from natural language descriptions.
// DALL·E 3 can create new images given a prompt, while DALL·E 2 also supports editing existing images
// and generating variations of user-provided images.
// Reference: https://platform.openai.com/docs/models#dall-e
const (
	// The latest DALL·E model, released in November 2023.
	DALL_E_3 Model = "dall-e-3"
	// The previous DALL·E model, released in November 2022. This second iteration significantly improves image realism,
	// accuracy, and resolution (4× greater) compared to the original model.
	DALL_E_2 Model = "dall-e-2"
)

func (m Model) String() string {
	return string(m)
}

func (m Model) IsEmpty() bool {
	return m == ""
}

func (m Model) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}

func (m *Model) UnmarshalText(text []byte) error {
	*m = Model(text)
	return nil
}

func (m Model) MarshalJSON() ([]byte, error) {
	return []byte(`"` + m.String() + `"`), nil
}

func (m *Model) UnmarshalJSON(data []byte) error {
	*m = Model(data)
	return nil
}
