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
		&GptOss120B{},
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
	case "gpt-oss-120b":
		return &GptOss120B{}, true
	}
}

// NewGenerativeModel creates a new instance of a generative model
func NewGenerativeModel(modelName, apiKey string) (assistant.GenerativeModel, error) {
	switch modelName {
	case "gpt-oss-120b":
		return NewGptOss120B(apiKey), nil
	}
	return nil, fmt.Errorf("unsupported model name: %s", modelName)
}
