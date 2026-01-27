package cerebras

import (
	"fmt"

	"micheam.com/aico/internal/assistant"
)

// Endpoint is the Cerebras API endpoint (OpenAI-compatible)
const Endpoint = "https://api.cerebras.ai/v1/chat/completions"

// ProviderName is the name of this provider
const ProviderName = "cerebras"

// AvailableModels returns a list of available models
func AvailableModels() []assistant.ModelDescriptor {
	return []assistant.ModelDescriptor{
		&Llama3_1_8B{},
		&GptOss120B{},
		&Qwen3_235B_A22B_Instruct_2507{},
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
	case "llama3.1-8b":
		return &Llama3_1_8B{}, true
	case "gpt-oss-120b":
		return &GptOss120B{}, true
	case "qwen-3-235b-a22b-instruct-2507":
		return &Qwen3_235B_A22B_Instruct_2507{}, true
	}
}

// NewGenerativeModel creates a new instance of a generative model
func NewGenerativeModel(modelName, apiKey string) (assistant.GenerativeModel, error) {
	switch modelName {
	case "llama3.1-8b":
		return NewLlama3_1_8B(apiKey), nil
	case "gpt-oss-120b":
		return NewGptOss120B(apiKey), nil
	case "qwen-3-235b-a22b-instruct-2507":
		return NewQwen3_235B_A22B_Instruct_2507(apiKey), nil
	}
	return nil, fmt.Errorf("unsupported model name: %s", modelName)
}
