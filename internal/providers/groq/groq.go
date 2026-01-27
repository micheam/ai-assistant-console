package groq

import (
	"fmt"

	"micheam.com/aico/internal/assistant"
)

// Endpoint is the Groq API endpoint (OpenAI-compatible)
const Endpoint = "https://api.groq.com/openai/v1/chat/completions"

// ProviderName is the name of this provider
const ProviderName = "groq"

// AvailableModels returns a list of available models
func AvailableModels() []assistant.ModelDescriptor {
	return []assistant.ModelDescriptor{
		&Llama3_3_70B{},
		&Llama3_1_8B{},
		&Mixtral8x7B{},
	}
}

func DescribeModel(modelName string) (desc string, found bool) {
	m, ok := selectModel(modelName)
	if !ok {
		return "", false
	}
	return m.Description(), true
}

func selectModel(modelName string) (assistant.GenerativeModel, bool) {
	switch modelName {
	default:
		return nil, false
	case "llama-3.3-70b-versatile":
		return &Llama3_3_70B{}, true
	case "llama-3.1-8b-instant":
		return &Llama3_1_8B{}, true
	case "mixtral-8x7b-32768":
		return &Mixtral8x7B{}, true
	}
}

// NewGenerativeModel creates a new instance of a generative model
func NewGenerativeModel(modelName, apiKey string) (assistant.GenerativeModel, error) {
	switch modelName {
	case "llama-3.3-70b-versatile":
		return NewLlama3_3_70B(apiKey), nil
	case "llama-3.1-8b-instant":
		return NewLlama3_1_8B(apiKey), nil
	case "mixtral-8x7b-32768":
		return NewMixtral8x7B(apiKey), nil
	}
	return nil, fmt.Errorf("unsupported model name: %s", modelName)
}
